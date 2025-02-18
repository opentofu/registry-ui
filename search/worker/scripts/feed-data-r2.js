// ls /tmp/registry
// # npx wrangler r2 object put registry-api/$filename --local --file $filename

// find /tmp/registry -exec npx wrangler r2 object put registry-api/$filename --local --file $filename {} \;
const fs = require('fs');
const path = require('path');
const { exec } = require('node:child_process');

const directoryPath = '/tmp/registry';

fs.readdir(directoryPath, { recursive: true }, (err, files) => {
  if (err) {
    console.error('Error reading directory:', err);
    return;
  }

  files.forEach(fileName => {
    const realFilePath = path.join(directoryPath, fileName);
    const r2Name = realFilePath.replace(directoryPath, "");

    const cmd = `npx wrangler r2 object put registry-ui-api${r2Name} --file ${realFilePath} --local`
    console.log(cmd)
    exec(cmd, function (err, stdout, stderr) {
      console.log(stdout);
      console.log(stderr);
    });

  });
});
