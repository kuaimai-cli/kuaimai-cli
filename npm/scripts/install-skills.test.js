const test = require("node:test");
const assert = require("node:assert/strict");
const fs = require("fs");
const os = require("os");
const path = require("path");

const {
  isValidBundledSkillsRoot,
  installBundledSkills,
  copyDirRecursive,
} = require("./install-skills");

test("isValidBundledSkillsRoot requires default skills", () => {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), "km-skills-"));
  assert.equal(isValidBundledSkillsRoot(dir), false);
  for (const name of ["kuaimai-shared", "kuaimai-item", "kuaimai-scm"]) {
    const skillDir = path.join(dir, name);
    fs.mkdirSync(skillDir, { recursive: true });
    fs.writeFileSync(path.join(skillDir, "SKILL.md"), `# ${name}\n`);
  }
  assert.equal(isValidBundledSkillsRoot(dir), true);
});

test("installBundledSkills copies into agent directories", () => {
  const tmp = fs.mkdtempSync(path.join(os.tmpdir(), "km-install-"));
  const bundled = path.join(tmp, "bundled");
  for (const name of ["kuaimai-shared", "kuaimai-item", "kuaimai-scm"]) {
    const skillDir = path.join(bundled, name);
    fs.mkdirSync(path.join(skillDir, "references"), { recursive: true });
    fs.writeFileSync(path.join(skillDir, "SKILL.md"), `# ${name}\n`);
    fs.writeFileSync(path.join(skillDir, "references", "demo.md"), "# demo\n");
  }

  const home = path.join(tmp, "home");
  fs.mkdirSync(home, { recursive: true });
  const oldHome = process.env.HOME;
  process.env.HOME = home;
  process.env.KUAIMAI_CLI_SKILLS_DIR = bundled;

  try {
    const result = installBundledSkills({ force: true });
    assert.equal(result.skipped, false);
    assert.equal(result.source, bundled);

    const itemRef = path.join(home, ".cursor", "skills", "kuaimai-item", "references", "demo.md");
    assert.equal(fs.existsSync(itemRef), true);
  } finally {
    process.env.HOME = oldHome;
    delete process.env.KUAIMAI_CLI_SKILLS_DIR;
  }
});

test("copyDirRecursive copies nested files", () => {
  const tmp = fs.mkdtempSync(path.join(os.tmpdir(), "km-copy-"));
  const src = path.join(tmp, "src", "nested");
  fs.mkdirSync(src, { recursive: true });
  fs.writeFileSync(path.join(src, "a.txt"), "a");
  const dest = path.join(tmp, "dest");
  copyDirRecursive(path.join(tmp, "src"), dest);
  assert.equal(fs.readFileSync(path.join(dest, "nested", "a.txt"), "utf8"), "a");
});
