import {execFileSync, spawnSync} from 'node:child_process';
import {cpSync, mkdirSync, readFileSync, rmSync} from 'node:fs';
import path from 'node:path';
import {fileURLToPath} from 'node:url';

import {maintainerDocExcludes} from './docs-excludes.mjs';

const siteDir = path.dirname(fileURLToPath(import.meta.url));
const repoDir = path.resolve(siteDir, '../..');
const docsDir = path.join(repoDir, 'docs');
const versionedDocsDir = path.join(siteDir, 'versioned_docs');
const versions = JSON.parse(
  readFileSync(path.join(siteDir, 'versions.json'), 'utf8'),
);

rmSync(versionedDocsDir, {recursive: true, force: true});
mkdirSync(versionedDocsDir, {recursive: true});

for (const version of versions) {
  if (!/^\d+\.\d+$/.test(version)) {
    throw new Error(`Invalid documentation version: ${version}`);
  }

  const destination = path.join(versionedDocsDir, `version-${version}`);
  mkdirSync(destination, {recursive: true});

  const tag = execFileSync(
    'git',
    ['tag', '--list', `v${version}.*`, '--sort=-v:refname'],
    {cwd: repoDir, encoding: 'utf8'},
  )
    .trim()
    .split('\n')[0];

  if (tag) {
    const archive = execFileSync('git', ['archive', tag, 'docs'], {
      cwd: repoDir,
    });
    const extracted = spawnSync(
      'tar',
      ['-x', '-C', destination, '--strip-components=1'],
      {input: archive},
    );
    if (extracted.status !== 0) {
      throw new Error(
        `Could not extract documentation for ${tag}: ${extracted.stderr}`,
      );
    }
    console.log(`Generated documentation version ${version} from ${tag}.`);
  } else {
    console.warn(
      `WARNING: No tag matching v${version}.*; using current docs/ for version ${version}.`,
    );
    cpSync(docsDir, destination, {recursive: true});
  }

  for (const excludedDoc of maintainerDocExcludes) {
    rmSync(path.join(destination, excludedDoc), {force: true});
  }
}
