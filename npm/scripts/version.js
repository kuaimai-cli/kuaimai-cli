"use strict";

/** @param {string} v */
function normalizeVer(v) {
  if (!v) return "";
  const m = String(v).match(/(\d+\.\d+\.\d+)/);
  return m ? m[1] : "";
}

/** Semver-ish compare: true if a < b */
function versionLess(a, b) {
  a = normalizeVer(a);
  b = normalizeVer(b);
  if (!a || a === "dev") return !!b;
  if (!b) return false;
  if (a === b) return false;
  const ap = a.split(".");
  const bp = b.split(".");
  for (let i = 0; i < Math.max(ap.length, bp.length); i++) {
    const da = Number(ap[i]) || 0;
    const db = Number(bp[i]) || 0;
    if (da < db) return true;
    if (da > db) return false;
  }
  return false;
}

module.exports = { normalizeVer, versionLess };
