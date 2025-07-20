import fs from 'fs';
import path from 'path';

type Config = {
  queriesDir: string;
  outputFile: string;
  exclude?: string[];
  camelCaseKeys?: boolean;
};

function camelCase(str: string): string {
  return str.replace(/[-_](\w)/g, (_, c) => c.toUpperCase());
}

function isExcluded(filePath: string, excludeList: string[]): boolean {
  return excludeList.some((ex) => filePath.includes(ex) || filePath.endsWith(ex));
}

async function generateQueryIndex() {
  const configPath = path.resolve('generate-queries.config.json');

  if (!fs.existsSync(configPath)) {
    throw new Error(`Missing config file: ${configPath}`);
  }

  const config: Config = JSON.parse(fs.readFileSync(configPath, 'utf8'));
  const { queriesDir, outputFile, exclude = [], camelCaseKeys = false } = config;

  const absQueriesDir = path.resolve(queriesDir);
  const absOutputFile = path.resolve(outputFile);
  const files = fs.readdirSync(absQueriesDir);

  const sqlFiles = files.filter(
    (file) => file.endsWith('.sql.ts') && !file.startsWith('generated-') && !isExcluded(file, exclude),
  );

  if (sqlFiles.length === 0) {
    console.warn('⚠️ No SQL files found for codegen.');
    return;
  }

  const imports: string[] = [];
  const objectEntries: string[] = [];
  const typeEntries: string[] = [];
  const declares: string[] = [];

  console.log(`📦 Generating queries from: ${absQueriesDir}`);
  console.log(`➡️ Output: ${absOutputFile}`);
  console.log(`🧹 Excluding: ${exclude.join(', ') || '(none)'}`);
  console.log(`\n✅ Included:`);

  for (const file of sqlFiles) {
    const baseName = path.basename(file, '.sql.ts');
    const varName = camelCaseKeys ? camelCase(baseName) : baseName;

    let importPath = path
      .relative(path.dirname(absOutputFile), path.join(absQueriesDir, file))
      .replace(/\\/g, '/')
      .replace(/\.ts$/, '.js');

    if (!importPath.startsWith('.')) {
      importPath = './' + importPath;
    }

    imports.push(`import * as ${varName} from '${importPath}';`);
    objectEntries.push(`  ${varName},`);
    typeEntries.push(`  ${varName}: typeof ${varName};`);
    declares.push(`  declare ${varName}: WrappedPreparedQueries<GeneratedQueryTypes['${varName}']>;`);

    console.log(` - ${file} as ${varName}`);
  }

  const output =
    `
/**
 * AUTO-GENERATED FILE — DO NOT EDIT MANUALLY
 * Run "npm run generate:queries" to regenerate.
 */

import { PreparedQuery } from '@pgtyped/runtime';
${imports.join('\n')}

export type WrappedPreparedQuery<P extends PreparedQuery<any, any>> =
  P extends PreparedQuery<infer Param, infer Result> ? (param: Param) => Promise<Result[]> : never;

export type WrappedPreparedQueries<T> = {
  [K in keyof T]: T[K] extends PreparedQuery<any, any> ? WrappedPreparedQuery<T[K]> : never;
};

export type GeneratedQueryTypes = {
${typeEntries.join('\n')}
};

export abstract class ServiceWithGeneratedQueries {
${declares.join('\n')}
}

export const generatedQueries: GeneratedQueryTypes = {
${objectEntries.join('\n')}
};
`.trim() + '\n';

  fs.writeFileSync(absOutputFile, output, 'utf8');
  console.log('\n✅ Query index generation complete.');
}

generateQueryIndex().catch((err) => {
  console.error('❌ Failed to generate query index:', err);
  process.exit(1);
});
