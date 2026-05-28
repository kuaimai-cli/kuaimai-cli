#!/usr/bin/env node

const fs = require("fs");
const path = require("path");
const { execFileSync } = require("child_process");
const os = require("os");
const crypto = require("crypto");

const pkg = require("../package.json");
const VERSION = String(pkg.binaryVersion || pkg.version).replace(/-.*$/, "");
const REPO = "kuaimai-cli/kuaimai-cli";
const NAME = "kuaimai-cli";
const DEFAULT_MIRROR_HOST = "https://registry.npmmirror.com";
const DOWNLOAD_MAX_TIME_SEC = "300";

const ALLOWED_HOSTS = new Set([
  "github.com",
  "objects.githubusercontent.com",
  "registry.npmmirror.com",
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
    return new URL(url).hostname === "registry.npmjs.org";
  } catch (_) {
    return false;
  }
}

function resolveMirrorUrls(env, archive, version) {
  const binaryPath = `/-/binary/kuaimai-cli/v${version}/${archive}`;
  const urls = [];
  const registry = (env.npm_config_registry || "").trim();
  if (registry && !isDefaultNpmjsRegistry(registry) && isValidDownloadBase(registry)) {
    const base = new URL(registry);
    urls.push(joinUrl(base.origin + base.pathname, binaryPath));
  }
  const mirror = (env.KUAIMAI_CLI_NPM_MIRROR || "").trim() || DEFAULT_MIRROR_HOST;
  if (env.KUAIMAI_CLI_USE_NPM_MIRROR === "1") {
    const mirrorUrl = joinUrl(mirror, binaryPath);
    if (!urls.includes(mirrorUrl)) urls.push(mirrorUrl);
  }
  return urls;
}

function assertAllowedHost(url) {
  const { hostname } = new URL(url);
  if (!ALLOWED_HOSTS.has(hostname)) {
    throw new Error(`Download host not allowed: ${hostname}`);
  }
}

function getMirrorUrls(env) {
  const urls = resolveMirrorUrls(env, archiveName, VERSION);
  for (const u of urls) ALLOWED_HOSTS.add(new URL(u).hostname);
  return urls;
}

function download(url, destPath) {
  assertAllowedHost(url);
  const args = [
    "--fail", "--location", "--silent", "--show-error",
    "--connect-timeout", "15", "--max-time", DOWNLOAD_MAX_TIME_SEC,
    "--max-redirs", "3",
    "--output", destPath,
  ];
  if (isWindows) args.unshift("--ssl-revoke-best-effort");
  args.push(url);
  execFileSync("curl", args, { stdio: ["ignore", "ignore", "pipe"] });
}

function extractZipWindows(archivePath, destDir) {
  const psOpts = ["-NoProfile", "-ExecutionPolicy", "Bypass", "-Command"];
  const psEnv = { ...process.env, KUAIMAI_CLI_ARCHIVE: archivePath, KUAIMAI_CLI_DEST: destDir };
  const dotnet =
    "$ErrorActionPreference='Stop';" +
    "Add-Type -AssemblyName System.IO.Compression.FileSystem;" +
    "[System.IO.Compression.ZipFile]::ExtractToDirectory($env:KUAIMAI_CLI_ARCHIVE,$env:KUAIMAI_CLI_DEST)";
  execFileSync("powershell.exe", [...psOpts, dotnet], { stdio: "inherit", env: psEnv });
}

function getExpectedChecksum(archive) {
  const checksumsPath = path.join(__dirname, "..", "checksums.txt");
  if (!fs.existsSync(checksumsPath)) {
    console.error("[WARN] checksums.txt not found, skipping checksum verification");
    return null;
  }
  const content = fs.readFileSync(checksumsPath, "utf8").trim();
  if (!content || content.startsWith("#")) return null;
  for (const line of content.split("\n")) {
    const trimmed = line.trim();
    if (!trimmed) continue;
    const idx = trimmed.indexOf(" ");
    if (idx === -1) continue;
    const hash = trimmed.slice(0, idx);
    const name = trimmed.slice(idx + 2);
    if (name === archive) return hash;
  }
  return null;
}

function verifyChecksum(archivePath, expectedHash) {
  if (!expectedHash) return;
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
    throw new Error(`Checksum mismatch: expected ${expectedHash}, got ${actual}`);
  }
}

function install() {
  if (!platform || !arch) {
    throw new Error(`Unsupported platform: ${process.platform}-${process.arch}`);
  }

  const mirrorUrls = getMirrorUrls(process.env);
  const downloadUrls = [GITHUB_URL, ...mirrorUrls];
  fs.mkdirSync(binDir, { recursive: true });

  const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "kuaimai-cli-"));
  const archivePath = path.join(tmpDir, archiveName);

  try {
    let lastErr;
    let downloaded = false;
    for (const url of downloadUrls) {
      try {
        download(url, archivePath);
        downloaded = true;
        break;
      } catch (e) {
        lastErr = e;
      }
    }
    if (!downloaded) {
      const tried = downloadUrls.join("\n  ");
      const detail = lastErr && lastErr.message ? lastErr.message : String(lastErr);
      throw new Error(`All download URLs failed:\n  ${tried}\nLast error: ${detail}`);
    }

    verifyChecksum(archivePath, getExpectedChecksum(archiveName));

    if (isWindows) {
      extractZipWindows(archivePath, tmpDir);
    } else {
      execFileSync("tar", ["-xzf", archivePath, "-C", tmpDir], { stdio: "ignore" });
    }

    const binaryName = NAME + (isWindows ? ".exe" : "");
    fs.copyFileSync(path.join(tmpDir, binaryName), dest);
    fs.chmodSync(dest, 0o755);
    console.log(`${NAME} v${VERSION} installed successfully`);
  } finally {
    fs.rmSync(tmpDir, { recursive: true, force: true });
  }
}

if (require.main === module) {
  const isNpxPostinstall =
    process.env.npm_command === "exec" && !process.env.KUAIMAI_CLI_RUN;

  if (isNpxPostinstall) {
    process.exit(0);
  }

  try {
    install();
  } catch (err) {
    console.error(`Failed to install ${NAME}:`, err.message);
    process.exit(1);
  }
}

module.exports = { install };
