#!/usr/bin/env node

const { execFileSync } = require("child_process");
const fs = require("fs");
const path = require("path");

const ext = process.platform === "win32" ? ".exe" : "";
const bin = path.join(__dirname, "..", "bin", "kuaimai-cli" + ext);

const args = process.argv.slice(2);
if (args[0] === "install") {
  require("./install-wizard.js");
} else {
  if (!fs.existsSync(bin)) {
    try {
      execFileSync(process.execPath, [path.join(__dirname, "install.js")], {
        stdio: "inherit",
        env: { ...process.env, KUAIMAI_CLI_RUN: "true" },
      });
    } catch (_) {
      console.error(`\nFailed to auto-install kuaimai-cli binary.\n`);
      process.exit(1);
    }
  }

  try {
    execFileSync(bin, args, { stdio: "inherit" });
  } catch (e) {
    process.exit(e.status || 1);
  }
}
