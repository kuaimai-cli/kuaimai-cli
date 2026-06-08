#!/usr/bin/env node

const fs = require("fs");
const path = require("path");

const repoSkills = path.join(__dirname, "..", "..", "skills");
const npmSkills = path.join(__dirname, "..", "skills");
const required = ["kuaimai-shared", "kuaimai-item", "kuaimai-scm"];

function copyDirRecursive(src, dest) {
  fs.mkdirSync(dest, { recursive: true });
  for (const entry of fs.readdirSync(src, { withFileTypes: true })) {
    const srcPath = path.join(src, entry.name);
    const destPath = path.join(dest, entry.name);
    if (entry.isDirectory()) {
      copyDirRecursive(srcPath, destPath);
    } else if (entry.isFile()) {
      fs.copyFileSync(srcPath, destPath);
    }
  }
}

function main() {
  if (!fs.existsSync(repoSkills)) {
    console.error(`skills 源目录不存在: ${repoSkills}`);
    process.exit(1);
  }
  for (const name of required) {
    const skillMD = path.join(repoSkills, name, "SKILL.md");
    if (!fs.existsSync(skillMD)) {
      console.error(`缺少 ${skillMD}`);
      process.exit(1);
    }
  }
  if (fs.existsSync(npmSkills)) {
    fs.rmSync(npmSkills, { recursive: true, force: true });
  }
  copyDirRecursive(repoSkills, npmSkills);
  console.log(`已同步 skills/ → ${npmSkills}`);
}

main();
