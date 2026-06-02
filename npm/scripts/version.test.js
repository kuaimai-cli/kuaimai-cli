"use strict";

const assert = require("assert");
const { normalizeVer, versionLess } = require("./version");

assert.strictEqual(normalizeVer("kuaimai-cli version 0.1.0 (2026-05-28)"), "0.1.0");
assert.strictEqual(versionLess("0.1.0", "0.1.7"), true);
assert.strictEqual(versionLess("0.1.7", "0.1.0"), false);
assert.strictEqual(versionLess("0.1.7", "0.1.7"), false);
console.log("version.test.js ok");
