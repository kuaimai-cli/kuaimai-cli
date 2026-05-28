#!/usr/bin/env node
"use strict";

const fs = require("fs");
const path = require("path");
const { execFileSync } = require("child_process");

function ensureExecutable(filePath) {
  if (!filePath || !fs.existsSync(filePath)) return;
  try {
    fs.chmodSync(filePath, 0o755);
  } catch (_) {
    // best effort
  }
}

function stripMacOSQuarantine(filePath) {
  if (process.platform !== "darwin" || !filePath || !fs.existsSync(filePath)) return;
  try {
    execFileSync("xattr", ["-d", "com.apple.quarantine", filePath], {
      stdio: "ignore",
    });
  } catch (_) {
    // attribute may not exist
  }
}

function ensurePackageEntrypoints(packageRoot) {
  const scriptsDir = path.join(packageRoot, "scripts");
  for (const name of ["run.js", "install.js", "install-wizard.js"]) {
    ensureExecutable(path.join(scriptsDir, name));
  }

  const ext = process.platform === "win32" ? ".exe" : "";
  const goBin = path.join(packageRoot, "bin", "kuaimai-cli" + ext);
  ensureExecutable(goBin);
  stripMacOSQuarantine(goBin);
}

module.exports = {
  ensureExecutable,
  stripMacOSQuarantine,
  ensurePackageEntrypoints,
};
