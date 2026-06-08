#!/usr/bin/env node

const fs = require("fs");
const path = require("path");
const { execFileSync, execFile } = require("child_process");
const p = require("@clack/prompts");

const PKG = "@kuaimai-cli/cli";
const TARGET_VERSION = require("../package.json").version.replace(/-.*$/, "");
const isWindows = process.platform === "win32";
const { ensureExecutable, ensurePackageEntrypoints } = require("./permissions");
const { normalizeVer, versionLess } = require("./version");
const { installBundledSkills } = require("./install-skills");

const messages = {
  zh: {
    setup: "正在设置 kuaimai-cli...",
    step1: "正在安装 %s...",
    step1Skip: "已是最新 (v%s)，跳过全局安装",
    step1Upgrade: "发现旧版 npm 包 (v%s)，正在升级到 v%s...",
    step1Done: "已全局安装",
    step1Fail: "全局安装失败。运行: npm install -g %s@%s",
    step2: "安装 AI Skills",
    step2Skip: "Skills 已安装，跳过",
    step2Spinner: "正在安装 Skills...",
    step2Done: "Skills 已安装",
    step2Fail: "Skills 安装失败。请重试 npx @kuaimai-cli/cli@latest install 或 kuaimai-cli skill install --force",
    step3: "正在初始化配置...",
    step3Skip: "跳过配置",
    step3Done: "配置已初始化",
    step3Fail: "配置失败。运行: kuaimai-cli config init",
    step4: "登录",
    step4Hint:
      'accessToken 须由用户提供。如尚未持有，请联系 ERP 管理员申请分配；拿到 token 后运行:\n  kuaimai-cli auth login "<accessToken>"',
    step4Skip: '跳过登录。后续向 ERP 管理员申请 accessToken 后运行 kuaimai-cli auth login "<accessToken>"',
    done:
      "安装完成！\n运行 kuaimai-cli --version 确认版本；kuaimai-cli auth status 验证登录。",
    cancelled: "安装已取消",
    nonTtyHint:
      "请在终端完成以下步骤:\n" +
      "  （accessToken 须由用户提供，可联系 ERP 管理员申请分配）\n" +
      '  kuaimai-cli auth login "<accessToken>"\n' +
      "  kuaimai-cli auth status --output json",
  },
};

function fmt(template, ...values) {
  let i = 0;
  return template.replace(/%s/g, () => values[i++] ?? "");
}

function forceInstall() {
  return process.env.KUAIMAI_CLI_FORCE_INSTALL === "1";
}

function execCmd(cmd, args, opts) {
  if (isWindows) {
    return execFileSync("cmd.exe", ["/c", cmd, ...args], opts);
  }
  return execFileSync(cmd, args, opts);
}

function run(cmd, args, opts = {}) {
  execCmd(cmd, args, { stdio: "inherit", ...opts });
}

function runSilent(cmd, args, opts = {}) {
  return execCmd(cmd, args, { stdio: ["ignore", "pipe", "pipe"], ...opts });
}

function runSilentAsync(cmd, args, opts = {}) {
  const actualCmd = isWindows ? "cmd.exe" : cmd;
  const actualArgs = isWindows ? ["/c", cmd, ...args] : args;
  return new Promise((resolve, reject) => {
    execFile(actualCmd, actualArgs, { stdio: ["ignore", "pipe", "pipe"], ...opts }, (err, stdout) => {
      if (err) reject(err);
      else resolve(stdout);
    });
  });
}

function globalPackageRoot() {
  try {
    const prefix = execFileSync("npm", ["prefix", "-g"], {
      stdio: ["ignore", "pipe", "pipe"],
    })
      .toString()
      .trim();
    return path.join(prefix, "lib", "node_modules", PKG);
  } catch (_) {
    return null;
  }
}

function goBinaryPath(root) {
  return path.join(root, "bin", "kuaimai-cli" + (isWindows ? ".exe" : ""));
}

function getBinaryVersion(binPath) {
  if (!binPath || !fs.existsSync(binPath)) return null;
  try {
    ensureExecutable(binPath);
    const out = runSilent(binPath, ["--version"], { timeout: 15000 });
    return normalizeVer(out.toString());
  } catch (_) {
    return null;
  }
}

function packageRoots() {
  const roots = [path.join(__dirname, "..")];
  const globalRoot = globalPackageRoot();
  if (globalRoot && !roots.includes(globalRoot)) roots.push(globalRoot);
  return roots;
}

function resolveCliRunner() {
  let best = null;
  let bestVer = "";

  for (const root of packageRoots()) {
    const candidate = goBinaryPath(root);
    if (!fs.existsSync(candidate)) continue;
    const ver = getBinaryVersion(candidate) || "";
    if (!best || versionLess(bestVer, ver)) {
      best = candidate;
      bestVer = ver;
    }
  }

  if (best) {
    ensureExecutable(best);
    return { cmd: best, argsPrefix: [] };
  }

  const runJs = path.join(__dirname, "run.js");
  ensureExecutable(runJs);
  return { cmd: process.execPath, argsPrefix: [runJs] };
}

function runCLI(args, opts = {}) {
  const { cmd, argsPrefix } = resolveCliRunner();
  return runSilent(cmd, [...argsPrefix, ...args], opts);
}

function getGloballyInstalledVersion() {
  try {
    const out = runSilent("npm", ["list", "-g", PKG], { timeout: 15000 });
    const match = out.toString().match(/@(\d+\.\d+\.\d+[^\s]*)/);
    return match ? normalizeVer(match[1]) : null;
  } catch (_) {
    return null;
  }
}

function ensurePackageBinary(packageRoot) {
  const installScript = path.join(packageRoot, "scripts", "install.js");
  if (!fs.existsSync(installScript)) return null;

  const bin = goBinaryPath(packageRoot);
  const ver = getBinaryVersion(bin);
  if (ver && !versionLess(ver, TARGET_VERSION) && !forceInstall()) {
    return bin;
  }
  execFileSync(process.execPath, [installScript], {
    stdio: "inherit",
    env: { ...process.env, KUAIMAI_CLI_RUN: "true" },
    cwd: packageRoot,
  });
  return bin;
}

function ensureAllBinaries() {
  for (const root of packageRoots()) {
    ensurePackageBinary(root);
  }
}

async function stepInstallGlobally(msg) {
  const installedVer = getGloballyInstalledVersion();
  const pkgSpec = `${PKG}@${TARGET_VERSION}`;

  if (
    installedVer &&
    !forceInstall() &&
    !versionLess(installedVer, TARGET_VERSION)
  ) {
    p.log.info(fmt(msg.step1Skip, installedVer));
    return false;
  }

  if (installedVer && versionLess(installedVer, TARGET_VERSION)) {
    p.log.info(fmt(msg.step1Upgrade, installedVer, TARGET_VERSION));
  }

  const s = p.spinner();
  s.start(fmt(msg.step1, pkgSpec));
  try {
    await runSilentAsync("npm", ["install", "-g", pkgSpec], { timeout: 120000 });
    s.stop(msg.step1Done);
    return true;
  } catch (_) {
    s.stop(fmt(msg.step1Fail, PKG, TARGET_VERSION));
    process.exit(1);
  }
}

async function stepInstallSkills(msg, { refreshedCLI } = {}) {
  const s = p.spinner();
  s.start(msg.step2Spinner);
  try {
    const force = refreshedCLI || forceInstall();
    const result = installBundledSkills({ force });
    if (result.skipped) {
      s.stop(msg.step2Skip);
      return;
    }
    s.stop(msg.step2Done);
  } catch (_) {
    s.stop(msg.step2Fail);
    process.exit(1);
  }
}

async function stepConfigInit(msg) {
  const s = p.spinner();
  s.start(msg.step3);
  try {
    runCLI(["config", "init"], { timeout: 15000 });
    s.stop(msg.step3Done);
  } catch (_) {
    s.stop(msg.step3Fail);
    p.log.warn(msg.step3Skip);
  }
}

async function stepAuthHint(msg) {
  p.log.step(msg.step4);
  p.log.info(msg.step4Hint);
}

async function runSetup(msg) {
  const refreshedCLI = await stepInstallGlobally(msg);
  for (const root of packageRoots()) {
    ensurePackageEntrypoints(root);
  }
  ensureAllBinaries();
  await stepInstallSkills(msg, { refreshedCLI: refreshedCLI || forceInstall() });
  await stepConfigInit(msg);
  await stepAuthHint(msg);
}

async function main() {
  const msg = messages.zh;
  const isInteractive = !!process.stdin.isTTY;

  if (isInteractive) {
    p.intro(msg.setup);
    await runSetup(msg);
    p.outro(msg.done);
  } else {
    console.log(msg.setup);
    await runSetup(msg);
    console.log(msg.nonTtyHint);
  }
}

main().catch((err) => {
  p.cancel("Unexpected error: " + (err.message || err));
  process.exit(1);
});
