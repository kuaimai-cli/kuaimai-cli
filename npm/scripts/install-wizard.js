#!/usr/bin/env node

const fs = require("fs");
const path = require("path");
const { execFileSync, execFile } = require("child_process");
const p = require("@clack/prompts");

const PKG = "@kuaimai/cli";
const isWindows = process.platform === "win32";

const messages = {
  zh: {
    setup: "正在设置 kuaimai-cli...",
    step1: "正在安装 %s...",
    step1Skip: "已安装 (v%s)，跳过",
    step1Done: "已全局安装",
    step1Fail: "全局安装失败。运行: npm install -g %s",
    step2: "安装 AI Skills",
    step2Skip: "Skills 已安装，跳过",
    step2Spinner: "正在安装 Skills...",
    step2Done: "Skills 已安装",
    step2Fail: "Skills 安装失败。运行: kuaimai-cli skill install-all",
    step3: "正在初始化配置...",
    step3Skip: "跳过配置",
    step3Done: "配置已初始化",
    step3Fail: "配置失败。运行: kuaimai-cli config init",
    step4: "登录",
    step4Hint: "请从 ERP 浏览器 DevTools → Network 复制 accessToken，然后运行:\n  kuaimai-cli auth login \"<accessToken>\"",
    step4Skip: "跳过登录。后续运行 kuaimai-cli auth login \"<accessToken>\"",
    done: "安装完成！\n运行 kuaimai-cli auth status 验证；Agent 请阅读 docs/kuaimai-cli-agent-installation-guide.md",
    cancelled: "安装已取消",
    nonTtyHint:
      "请在终端完成以下步骤:\n" +
      "  kuaimai-cli config init\n" +
      "  kuaimai-cli auth login \"<accessToken>\"\n" +
      "  kuaimai-cli auth status --output json",
  },
};

function fmt(template, ...values) {
  let i = 0;
  return template.replace(/%s/g, () => values[i++] ?? "");
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

function whichKuaimaiCli() {
  try {
    const prefix = execFileSync("npm", ["prefix", "-g"], {
      stdio: ["ignore", "pipe", "pipe"],
    })
      .toString()
      .trim();
    const candidate = isWindows
      ? path.join(prefix, "kuaimai-cli.cmd")
      : path.join(prefix, "bin", "kuaimai-cli");
    if (fs.existsSync(candidate)) return candidate;
  } catch (_) {
    // fall through
  }
  try {
    const cmd = isWindows ? "where" : "which";
    return execFileSync(cmd, ["kuaimai-cli"], { stdio: ["ignore", "pipe", "pipe"] })
      .toString()
      .split("\n")[0]
      .trim();
  } catch (_) {
    return null;
  }
}

function getGloballyInstalledVersion() {
  try {
    const out = runSilent("npm", ["list", "-g", PKG], { timeout: 15000 });
    const match = out.toString().match(/@(\d+\.\d+\.\d+[^\s]*)/);
    return match ? match[1] : "unknown";
  } catch (_) {
    return null;
  }
}

function handleCancel(value, msg) {
  if (p.isCancel(value)) {
    p.cancel(msg.cancelled);
    process.exit(0);
  }
  return value;
}

async function stepInstallGlobally(msg) {
  const installedVer = getGloballyInstalledVersion();
  if (installedVer) {
    p.log.info(fmt(msg.step1Skip, installedVer));
    return;
  }
  const s = p.spinner();
  s.start(fmt(msg.step1, PKG));
  try {
    await runSilentAsync("npm", ["install", "-g", PKG], { timeout: 120000 });
    s.stop(msg.step1Done);
  } catch (_) {
    s.stop(fmt(msg.step1Fail, PKG));
    process.exit(1);
  }
}

async function skillsAlreadyInstalled(binPath) {
  try {
    const out = runSilent(binPath, ["skill", "list", "--output", "json"], { timeout: 30000 });
    return /kuaimai-item/.test(out.toString());
  } catch (_) {
    return false;
  }
}

async function stepInstallSkills(msg, binPath) {
  const s = p.spinner();
  s.start(msg.step2Spinner);
  try {
    if (await skillsAlreadyInstalled(binPath)) {
      s.stop(msg.step2Skip);
      return;
    }
    runSilent(binPath, ["skill", "install-all"], { timeout: 120000 });
    s.stop(msg.step2Done);
  } catch (_) {
    s.stop(msg.step2Fail);
    process.exit(1);
  }
}

async function stepConfigInit(msg, binPath) {
  const s = p.spinner();
  s.start(msg.step3);
  try {
    runSilent(binPath, ["config", "init"], { timeout: 15000 });
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

async function main() {
  const msg = messages.zh;
  const isInteractive = !!process.stdin.isTTY;

  if (isInteractive) {
    p.intro(msg.setup);
    await stepInstallGlobally(msg);

    let binPath = whichKuaimaiCli();
    if (!binPath) {
      binPath = path.join(__dirname, "..", "bin", "kuaimai-cli" + (isWindows ? ".exe" : ""));
      if (!fs.existsSync(binPath)) {
        execFileSync(process.execPath, [path.join(__dirname, "install.js")], {
          stdio: "inherit",
          env: { ...process.env, KUAIMAI_CLI_RUN: "true" },
        });
      }
    }

    await stepInstallSkills(msg, binPath);
    await stepConfigInit(msg, binPath);
    await stepAuthHint(msg);
    p.outro(msg.done);
  } else {
    console.log(msg.setup);
    await stepInstallGlobally(msg);
    const binPath = whichKuaimaiCli();
    if (binPath) {
      try {
        runSilent(binPath, ["skill", "install-all"], { timeout: 120000 });
      } catch (_) {
        // best effort
      }
      try {
        runSilent(binPath, ["config", "init"], { timeout: 15000 });
      } catch (_) {
        // best effort
      }
    }
    console.log(msg.nonTtyHint);
  }
}

main().catch((err) => {
  p.cancel("Unexpected error: " + (err.message || err));
  process.exit(1);
});
