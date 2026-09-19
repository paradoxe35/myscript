// Embeds the Wit.ai keys from WITAI_KEYS into a generated Go source, so the
// keys never touch the repo. A no-op when WITAI_KEYS is unset: the app then
// compiles with zero keys and the Wit.ai option stays hidden.
const fs = require("fs");
const path = require("path");

const output = path.join(
  __dirname,
  "..",
  "internal",
  "transcribe",
  "wait.ai",
  "keys_embedded.go"
);

const keys = process.env.WITAI_KEYS;

if (!keys) {
  console.log("WITAI_KEYS not set; skipping the Wit.ai key embed");
  process.exit(0);
}

try {
  JSON.parse(keys);
} catch (error) {
  console.error(`WITAI_KEYS is not valid JSON: ${error.message}`);
  console.error('Expected [{"key":"...","lang":"en"}, ...]');
  process.exit(1);
}

// JSON.stringify escaping is a valid Go string literal: the quote, backslash
// and unicode rules match.
fs.writeFileSync(
  output,
  `package witai\n\nfunc init() {\n\trawKeys = ${JSON.stringify(keys)}\n}\n`,
  { encoding: "utf8" }
);

console.log(`Embedded the Wit.ai keys into ${output}`);
