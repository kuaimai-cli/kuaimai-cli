#!/usr/bin/env node

const fs = require("fs");
const os = require("os");
const path = require("path");

const PKG = "@kuaimai-cli/cli";
const DEFAULT_SKILL_NAMES = ["kuaimai-shared", "kuaimai-erp-item", "kuaimai-scm-item"];
const LEGACY_DEFAULT_SKILL_NAMES = ["kuaimai-item", "kuaimai-scm"];

function agentSkillRoots() {
  const home = os.homedir();
  return [
    path.join(home, ".agents", "skills"),
    path.join(home, ".cursor", "skills"),
    path.join(home, ".codex", "skills"),
    path.join(home, ".claude", "skills"),
    path.join(home, ".windsurf", "skills"),
  ];
}

function globalPackageRoot() {
  try {
    const { execFileSync } = require("child_process");
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

function packageRoots() {
  const roots = [path.join(__dirname, "..")];
  const globalRoot = globalPackageRoot();
  if (globalRoot && !roots.includes(globalRoot)) roots.push(globalRoot);
  return roots;
}

function isValidBundledSkillsRoot(root) {
  if (!root || !fs.existsSync(root)) return false;
  return DEFAULT_SKILL_NAMES.every((name) =>
    fs.existsSync(path.join(root, name, "SKILL.md"))
  );
}

function resolveBundledSkillsRoot() {
  const envDir = (process.env.KUAIMAI_CLI_SKILLS_DIR || "").trim();
  if (envDir && isValidBundledSkillsRoot(envDir)) {
    return envDir;
  }
  for (const root of packageRoots()) {
    const skills = path.join(root, "skills");
    if (isValidBundledSkillsRoot(skills)) {
      return skills;
    }
  }
  return null;
}

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

function skillsAlreadyInstalled() {
  return agentSkillRoots().every((root) =>
    DEFAULT_SKILL_NAMES.every((name) => {
      const skillMD = path.join(root, name, "SKILL.md");
      const refDir = path.join(root, name, "references");
      return fs.existsSync(skillMD) && fs.existsSync(refDir);
    })
  );
}

function installBundledSkills({ names = DEFAULT_SKILL_NAMES, force = false } = {}) {
 if (!force && skillsAlreadyInstalled()) {
    return { skipped: true, source: null, names: [] };
  }
  const bundled = resolveBundledSkillsRoot();
  if (!bundled) {
    throw new Error(
      "npm 包内未找到 skills/（请重新安装 @kuaimai-cli/cli 或设置 KUAIMAI_CLI_SKILLS_DIR）"
    );
  }
  for (const name of names) {
    const src = path.join(bundled, name);
    if (!fs.existsSync(path.join(src, "SKILL.md"))) {
      throw new Error(`bundled skill 缺少 SKILL.md: ${name}`);
    }
  }
  for (const root of agentSkillRoots()) {
    fs.mkdirSync(root, { recursive: true });
    if (isDefaultSkillSet(names)) {
      for (const name of [...DEFAULT_SKILL_NAMES, ...LEGACY_DEFAULT_SKILL_NAMES]) {
        fs.rmSync(path.join(root, name), { recursive: true, force: true });
      }
    }
    for (const name of names) {
      const src = path.join(bundled, name);
      const dest = path.join(root, name);
      fs.rmSync(dest, { recursive: true, force: true });
      copyDirRecursive(src, dest);
    }
  }
  return { skipped: false, source: bundled, names };
}

function isDefaultSkillSet(names) {
  if (!Array.isArray(names) || names.length !== DEFAULT_SKILL_NAMES.length) return false;
  return DEFAULT_SKILL_NAMES.every((name) => names.includes(name));
}

module.exports = {
  DEFAULT_SKILL_NAMES,
  LEGACY_DEFAULT_SKILL_NAMES,
  agentSkillRoots,
  packageRoots,
  resolveBundledSkillsRoot,
  skillsAlreadyInstalled,
  installBundledSkills,
  copyDirRecursive,
  isValidBundledSkillsRoot,
};
