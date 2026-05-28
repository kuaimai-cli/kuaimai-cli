#!/usr/bin/env node
// 对标 larksuite/cli/scripts/install.js：npm 薄壳 + postinstall 从 GitHub Release 下载 Go 二进制。
// 差异：暂不启用 npmmirror / registry 二进制镜像（须维护者同步后才可启用，否则 404）。

const fs = require("fs");
const path = require("path");
const { execFileSync } = require("child_process");
const os = require("os");
const crypto = require("crypto");

const VERSION = require("../package.json").version.replace(/-.*$/, "");
const REPO = "kuaimai-cli/kuaimai-cli";
const NAME = "kuaimai-cli";

// curl --location 会跟随重定向到 objects.githubusercontent.com
const ALLOWED_HOSTS = new Set([
  "github.com",
  "objects.githubusercontent.com",
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

function getDownloadUrl(env) {
  const override = (env.KUAIMAI_CLI_DOWNLOAD_URL || "").trim();
  return override || GITHUB_URL;
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
  if (isWindows) args.unshift("--ssl-revoke-best-effort");
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
            `.NET ZipFile: ${primaryErr.message}; ` +
            `Expand-Archive: ${secondErr.message}; ` +
            `tar: ${fallbackErr.message}`
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
    const name = trimmed.slice(idx + 2);
    if (name === archive) return hash;
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

function install() {
  if (!platform || !arch) {
    throw new Error(`Unsupported platform: ${process.platform}-${process.arch}`);
  }

  const downloadUrl = getDownloadUrl(process.env);
  fs.mkdirSync(binDir, { recursive: true });

  const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "kuaimai-cli-"));
  const archivePath = path.join(tmpDir, archiveName);

  try {
    download(downloadUrl, archivePath);
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
    console.log(`${NAME} v${VERSION} installed successfully`);
  } finally {
    fs.rmSync(tmpDir, { recursive: true, force: true });
  }
}

function formatInstallError(err, downloadUrl) {
  const releasePage = `https://github.com/${REPO}/releases/tag/v${VERSION}`;
  return [
    err && err.message ? err.message : "download failed",
    "",
    `Source: GitHub Release only (same as @larksuite/cli, without npmmirror fallback)`,
    `URL: ${downloadUrl}`,
    "",
    "If you are behind a firewall or in a restricted network, try one of:",
    "  # 1. Use a proxy:",
    "  export https_proxy=http://your-proxy:port",
    "  npm install -g @kuaimai-cli/cli",
    "",
    "  # 2. Manual install from Release:",
    `  open ${releasePage}`,
    `  download ${archiveName}, extract, and put kuaimai-cli on your PATH`,
    "",
    "  # 3. Point install.js at a mirror you control:",
    `  export KUAIMAI_CLI_DOWNLOAD_URL="<url-to-${archiveName}>"`,
    "  npm install -g @kuaimai-cli/cli",
  ].join("\n");
}

if (require.main === module) {
  // npx … install 向导不需要二进制；run.js 会以 KUAIMAI_CLI_RUN=1 触发下载
  const isNpxPostinstall =
    process.env.npm_command === "exec" && !process.env.KUAIMAI_CLI_RUN;

  if (isNpxPostinstall) {
    process.exit(0);
  }

  try {
    install();
  } catch (err) {
    console.error(`Failed to install ${NAME}:\n${formatInstallError(err, getDownloadUrl(process.env))}`);
    process.exit(1);
  }
}

module.exports = {
  install,
  GITHUB_URL,
  archiveName,
  getExpectedChecksum,
  verifyChecksum,
};
