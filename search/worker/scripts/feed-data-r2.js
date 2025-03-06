import { exec } from 'node:child_process';
import fs from 'fs';
import path from 'path';
import { promisify } from 'util';

const stat = promisify(fs.stat);
const pexec = promisify(exec);
const readdir = promisify(fs.readdir);

// This script is used to retrieve data generated by `go run generate.go` at backend folder.
// It will generate registry data that will be used by this script to feed into wrangler dev folder, mimicking a R2 bucket.
// This script expects data will be at /tmp/registry.
const DIR_PATH = '/tmp/registry';

const putObject = (directoryPath) => async (filename, i) => {
  const realFilePath = path.join(directoryPath, filename);
  const r2Name = realFilePath.replace(directoryPath, "");

  try {
    const stats = await stat(realFilePath)
    // Directories are not supported by `wrangler r2` API and
    // not needed since wrangler create the path without using folders.
    if(stats.isDirectory()) return
  } catch(err) {
    console.error(filename, err);
    return;
  }

  await runCmd(r2Name, realFilePath);
  return realFilePath;
};

const runCmd = async (bucketFileName, path) => {
  const cmd = `npx wrangler r2 object put registry-ui-api${bucketFileName} --file ${path} --local`
  try {
    await pexec(cmd)
  } catch(err) {
    console.error(path, err)
  }

  console.log(`Processed: ${path}`)
};

const run = async () => {
  console.log(`Reading ${DIR_PATH}...`)
  const files = await readdir(DIR_PATH, { recursive: true });
  const total = files.length;

  console.log(`Processing about ${total} files/folders...`);
  // We're processing sequentially because wranger r2 put doesn't handle
  // well concurrency giving a lot of errors, even with 2 files at once.
  for(const f of files) {
    await putObject(DIR_PATH)(f);
  }

  console.log("Processed everything. Search is not affected by this program, but now if you try to access opentofu/ad (default one on Makefile), it will be possible to see its details.")
}

await run()
