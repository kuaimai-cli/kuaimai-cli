"use strict";

const assert = require("assert");
const { getDownloadUrlChain, resolveMirrorUrls, GITHUB_URL, archiveName } = require("./install");

const archive = archiveName;
const version = require("../package.json").version.replace(/-.*$/, "");

assert.ok(GITHUB_URL.includes("github.com"));
assert.ok(GITHUB_URL.includes(archive));

const mirrors = resolveMirrorUrls({}, archive, version);
assert.strictEqual(
  mirrors[mirrors.length - 1],
  `https://registry.npmmirror.com/-/binary/kuaimai-cli/v${version}/${archive}`
);

const chain = getDownloadUrlChain({});
assert.strictEqual(chain[0], GITHUB_URL);
assert.ok(chain.length >= 2);
assert.ok(chain[1].includes("npmmirror.com"));

assert.strictEqual(resolveMirrorUrls({ KUAIMAI_CLI_SKIP_MIRROR: "1" }, archive, version).length, 0);

const override = getDownloadUrlChain({
  KUAIMAI_CLI_DOWNLOAD_URL: "https://cdn.npmmirror.com/binaries/kuaimai-cli/test.tar.gz",
});
assert.strictEqual(override.length, 1);

console.log("install-mirror.test.js ok");
