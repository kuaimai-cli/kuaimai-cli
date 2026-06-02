#!/usr/bin/env node
// 对标 @larksuite/cli/scripts/install.js：GitHub Release 优先，失败则回退 npmmirror /-/binary/ 镜像。

const fs = require("fs");
const path = require("path");
const { execFileSync } = require("child_process");
const os = require("os");
const crypto = require("crypto");

const VERSION = require("../package.json").version.replace(/-.*$/, "");
const REPO = "kuaimai-cli/kuaimai-cli";
const NAME = "kuaimai-cli";
const BINARY_PKG = "kuaimai-cli";
const DEFAULT_MIRROR_HOST = "https://registry.npmmirror.com";

// Allowlist gates the *initial* request URL only. curl --location follows redirects
// (capped by --max-redirs 3) without re-checking the target host.
const ALLOWED_HOSTS = new Set([
  "github.com",
  "objects.githubusercontent.com",
  "registry.npmmirror.com",
  "cdn.npmmirror.com",
]);

const PLATFORM_MAP = {
  darwin: "darwin",
  linux: "linux",
  win32: "windows",
};

const ARCH_MAP = {
  x64: "amd64",
  arm64: "arm64",
};

const platform = PLATFORM_MAP[process.platform];
const arch = ARCH_MAP[process.arch];
const isWindows = process.platform === "win32";
const ext = isWindows ? ".zip" : ".tar.gz";
const archiveName = `${NAME}-${VERSION}-${platform}-${arch}${ext}`;
const GITHUB_URL = `https://github.com/${REPO}/releases/download/v${VERSION}/${archiveName}`;

const binDir = path.join(__dirname, "..", "bin");
const dest = path.join(binDir, NAME + (isWindows ? ".exe" : ""));
const { ensurePackageEntrypoints, ensureExecutable, stripMacOSQuarantine } = require("./permissions");

function joinUrl(base, suffix) {
  return base.replace(/\/+$/, "") + suffix;
}

function isValidDownloadBase(raw) {
  try {
    const parsed = new URL(raw);
    return parsed.protocol === "https:" && !!parsed.hostname;
  } catch (_) {
    return false;
  }
}

function isDefaultNpmjsRegistry(url) {
  try {
    const { hostname } = new URL(url);
    return hostname === "registry.npmjs.org";
  } catch (_) {
    return false;
  }
}

/** @param {NodeJS.ProcessEnv} env */
function resolveMirrorUrls(env, archive, version) {
  if (env.KUAIMAI_CLI_SKIP_MIRROR === "1") return [];

  const binaryPath = `/-/binary/${BINARY_PKG}/v${version}/${archive}`;
  const defaultUrl = joinUrl(DEFAULT_MIRROR_HOST, binaryPath);

  const urls = [];
  const registry = (env.npm_config_registry || "").trim();
  if (registry && !isDefaultNpmjsRegistry(registry) && isValidDownloadBase(registry)) {
    const base = new URL(registry);
    urls.push(joinUrl(base.origin + base.pathname, binaryPath));
  }
  if (!urls.includes(defaultUrl)) urls.push(defaultUrl);
  return urls;
}

function assertAllowedHost(url) {
  const { hostname } = new URL(url);
  if (ALLOWED_HOSTS.has(hostname)) return;
  if (envAllowHost(hostname)) return;
  throw new Error(`Download host not allowed: ${hostname}`);
}

function envAllowHost(hostname) {
  return process.env.KUAIMAI_CLI_ALLOW_DOWNLOAD_HOST === hostname;
}

/** @param {NodeJS.ProcessEnv} env */
function getDownloadUrlChain(env) {
  const override = (env.KUAIMAI_CLI_DOWNLOAD_URL || "").trim();
  if (override) {
    assertAllowedHost(override);
    return [override];
  }

  const mirrorUrls = resolveMirrorUrls(env, archiveName, VERSION);
  for (const u of mirrorUrls) {
    try {
      ALLOWED_HOSTS.add(new URL(u).hostname);
    } catch (_) {
      /* ignore malformed mirror */
    }
  }
  return [GITHUB_URL, ...mirrorUrls];
}

function isCurlVersionSupported(versionOutput) {
  const match = String(versionOutput).match(/^\s*curl\s+(\d+)\.(\d+)\.(\d+)/i);
  if (!match) return false;
  const major = parseInt(match[1], 10);
  const minor = parseInt(match[2], 10);
  return major > 7 || (major === 7 && minor >= 70);
}

let _curlSupportsSslRevokeBestEffort;

function curlSupportsSslRevokeBestEffort() {
  if (_curlSupportsSslRevokeBestEffort !== undefined) {
    return _curlSupportsSslRevokeBestEffort;
  }
  try {
    const output = execFileSync("curl", ["--version"], {
      stdio: ["ignore", "pipe", "ignore"],
      encoding: "utf8",
      timeout: 5000,
    });
    _curlSupportsSslRevokeBestEffort = isCurlVersionSupported(output);
  } catch (_) {
    _curlSupportsSslRevokeBestEffort = false;
  }
  return _curlSupportsSslRevokeBestEffort;
}

function download(url, destPath) {
  assertAllowedHost(url);
  const args = [
    "--fail",
    "--location",
    "--silent",
    "--show-error",
    "--connect-timeout",
    "10",
    "--max-time",
    "120",
    "--max-redirs",
    "3",
    "--output",
    destPath,
  ];
  if (isWindows && curlSupportsSslRevokeBestEffort()) {
    args.unshift("--ssl-revoke-best-effort");
  }
  args.push(url);
  execFileSync("curl", args, { stdio: ["ignore", "ignore", "pipe"] });
}

function extractZipWindows(archivePath, destDir) {
  const psOpts = ["-NoProfile", "-ExecutionPolicy", "Bypass", "-Command"];
  const psStdio = ["ignore", "inherit", "inherit"];
  const psEnv = {
    ...process.env,
    KUAIMAI_CLI_ARCHIVE: archivePath,
    KUAIMAI_CLI_DEST: destDir,
  };

  try {
    const dotnet =
      "$ErrorActionPreference='Stop';" +
      "Add-Type -AssemblyName System.IO.Compression.FileSystem;" +
      "[System.IO.Compression.ZipFile]::ExtractToDirectory($env:KUAIMAI_CLI_ARCHIVE,$env:KUAIMAI_CLI_DEST)";
    execFileSync("powershell.exe", [...psOpts, dotnet], { stdio: psStdio, env: psEnv });
  } catch (primaryErr) {
    try {
      const cmdlet =
        "$ErrorActionPreference='Stop';" +
        "Expand-Archive -LiteralPath $env:KUAIMAI_CLI_ARCHIVE -DestinationPath $env:KUAIMAI_CLI_DEST -Force";
      execFileSync("powershell.exe", [...psOpts, cmdlet], { stdio: psStdio, env: psEnv });
    } catch (secondErr) {
      try {
        execFileSync("tar", ["-xf", archivePath, "-C", destDir], { stdio: psStdio });
      } catch (fallbackErr) {
        throw new Error(
          `Failed to extract ${archivePath}. ` +
            `ZipFile: ${primaryErr.message}; Expand-Archive: ${secondErr.message}; tar: ${fallbackErr.message}`
        );
      }
    }
  }
}

function getExpectedChecksum(archive) {
  const checksumsPath = path.join(__dirname, "..", "checksums.txt");
  if (!fs.existsSync(checksumsPath)) {
    console.error("[WARN] checksums.txt not found, skipping checksum verification");
    return null;
  }

  const content = fs.readFileSync(checksumsPath, "utf8");
  for (const line of content.split("\n")) {
    const trimmed = line.trim();
    if (!trimmed) continue;
    const idx = trimmed.indexOf(" ");
    if (idx === -1) continue;
    const hash = trimmed.slice(0, idx);
    const fileName = trimmed.slice(idx + 2);
    if (fileName === archive) return hash;
  }

  throw new Error(`Checksum entry not found for ${archive}`);
}

function verifyChecksum(archivePath, expectedHash) {
  if (expectedHash === null) return;

  const hash = crypto.createHash("sha256");
  const fd = fs.openSync(archivePath, "r");
  try {
    const buf = Buffer.alloc(64 * 1024);
    let bytesRead;
    while ((bytesRead = fs.readSync(fd, buf, 0, buf.length, null)) > 0) {
      hash.update(buf.subarray(0, bytesRead));
    }
  } finally {
    fs.closeSync(fd);
  }
  const actual = hash.digest("hex");
  if (actual.toLowerCase() !== expectedHash.toLowerCase()) {
    throw new Error(
      `[SECURITY] Checksum mismatch for ${path.basename(archivePath)}: expected ${expectedHash} but got ${actual}`
    );
  }
}

function downloadWithFallback(urls) {
  let lastErr;
  for (const url of urls) {
    const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "kuaimai-cli-"));
    const archivePath = path.join(tmpDir, archiveName);
    try {
      download(url, archivePath);
      return { archivePath, tmpDir, sourceUrl: url };
    } catch (e) {
      lastErr = e;
      fs.rmSync(tmpDir, { recursive: true, force: true });
    }
  }
  throw lastErr;
}

function install() {
  if (!platform || !arch) {
    throw new Error(`Unsupported platform: ${process.platform}-${process.arch}`);
  }

  const downloadUrls = getDownloadUrlChain(process.env);
  fs.mkdirSync(binDir, { recursive: true });

  const { archivePath, tmpDir, sourceUrl } = downloadWithFallback(downloadUrls);

  try {
    verifyChecksum(archivePath, getExpectedChecksum(archiveName));

    if (isWindows) {
      extractZipWindows(archivePath, tmpDir);
    } else {
      execFileSync("tar", ["-xzf", archivePath, "-C", tmpDir], { stdio: "ignore" });
    }

    const binaryName = NAME + (isWindows ? ".exe" : "");
    fs.copyFileSync(path.join(tmpDir, binaryName), dest);
    ensureExecutable(dest);
    stripMacOSQuarantine(dest);
    ensurePackageEntrypoints(path.join(__dirname, ".."));

    const viaMirror = sourceUrl !== GITHUB_URL && !process.env.KUAIMAI_CLI_DOWNLOAD_URL;
    const mirrorNote = viaMirror ? " (mirror)" : "";
    console.log(`${NAME} v${VERSION} installed successfully${mirrorNote}`);
  } finally {
    fs.rmSync(tmpDir, { recursive: true, force: true });
  }
}

function formatInstallError(err, downloadUrls) {
  const releasePage = `https://github.com/${REPO}/releases/tag/v${VERSION}`;
  const urls = Array.isArray(downloadUrls) ? downloadUrls : [downloadUrls];
  return [
    err && err.message ? err.message : "download failed",
    "",
    "Tried (in order):",
    ...urls.map((u) => `  - ${u}`),
    "",
    "If you are behind a firewall or in a restricted network, try one of:",
    "  # 1. Use a proxy:",
    "  export https_proxy=http://your-proxy:port",
    "  npm install -g @kuaimai-cli/cli",
    "",
    "  # 2. Local build from source:",
    "  git clone https://github.com/kuaimai-cli/kuaimai-cli.git && cd kuaimai-cli",
    "  make build && cp ./kuaimai-cli $(npm prefix -g)/bin/",
    "",
    "  # 3. Manual install from Release:",
    `  open ${releasePage}`,
    `  download ${archiveName}, extract, and put kuaimai-cli on your PATH`,
    "",
    "  # 4. Custom download URL:",
    `  export KUAIMAI_CLI_DOWNLOAD_URL="<url-to-${archiveName}>"`,
    "  npm install -g @kuaimai-cli/cli",
    "",
    "  # 5. npmmirror binary sync (maintainers): see docs/npmmirror-二进制镜像.md",
  ].join("\n");
}

if (require.main === module) {
  const isNpxPostinstall =
    process.env.npm_command === "exec" && !process.env.KUAIMAI_CLI_RUN;

  if (isNpxPostinstall) {
    process.exit(0);
  }

  const downloadUrls = getDownloadUrlChain(process.env);
  try {
    install();
  } catch (err) {
    console.error(`Failed to install ${NAME}:\n${formatInstallError(err, downloadUrls)}`);
    process.exit(1);
  }
}

module.exports = {
  install,
  GITHUB_URL,
  archiveName,
  getDownloadUrlChain,
  resolveMirrorUrls,
  getExpectedChecksum,
  verifyChecksum,
};
