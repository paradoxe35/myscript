const fs = require("fs");

// Get base64 string from command line argument
const base64Input = process.argv[2];

if (!base64Input) {
  console.error("Please provide a base64 string as argument");
  console.log("Usage: node script.js <base64string>");
  process.exit(1);
}

try {
  // Decode base64 to UTF-8 string
  const decodedString = Buffer.from(base64Input, "base64").toString("utf8");

  // keys_default.go covers builds without keys; the tag swaps it for this file.
  const tagged = decodedString.startsWith("//go:build")
    ? decodedString
    : "//go:build witai_keys\n\n" + decodedString;
  fs.writeFileSync("keys.go", tagged, { encoding: "utf8" });

  console.log("Successfully wrote decoded content to keys.go");
} catch (error) {
  console.error("Error:", error.message);
  process.exit(1);
}
