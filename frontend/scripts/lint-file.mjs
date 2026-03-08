import fs from 'node:fs';
import path from 'node:path';
import { ESLint } from 'eslint';
import ts from 'typescript';

const rootDir = process.cwd();
const srcDir = path.join(rootDir, 'src');
const defaultTargets = [
  path.join(srcDir, 'pages'),
  path.join(srcDir, 'hooks', 'useTheme.ts'),
  path.join(srcDir, 'hooks', 'useWailsEvent.ts'),
  path.join(srcDir, 'hooks', 'features'),
];
const sourceExtensions = new Set(['.ts', '.tsx']);
const ignoredSuffixes = ['.d.ts', '.test.ts', '.test.tsx'];

function isIgnoredFile(filePath) {
  return ignoredSuffixes.some((suffix) => filePath.endsWith(suffix));
}

function walkFiles(entryPath, results = []) {
  if (!fs.existsSync(entryPath)) {
    return results;
  }

  const stat = fs.statSync(entryPath);
  if (stat.isDirectory()) {
    for (const child of fs.readdirSync(entryPath)) {
      walkFiles(path.join(entryPath, child), results);
    }
    return results;
  }

  if (!sourceExtensions.has(path.extname(entryPath)) || isIgnoredFile(entryPath)) {
    return results;
  }

  results.push(path.normalize(entryPath));
  return results;
}

function collectTargetFiles(argv) {
  const rawTargets = argv.length > 0
    ? argv.map((value) => path.resolve(rootDir, value))
    : defaultTargets;
  const collected = new Set();

  for (const target of rawTargets) {
    for (const filePath of walkFiles(target)) {
      collected.add(filePath);
    }
  }

  return Array.from(collected).sort();
}

function collectSourceFiles() {
  return walkFiles(srcDir).sort();
}

function normalizeResolvedFile(resolvedPath) {
  if (!resolvedPath) {
    return null;
  }

  const normalized = path.normalize(resolvedPath);
  if (fs.existsSync(normalized) && fs.statSync(normalized).isFile()) {
    return normalized;
  }

  const candidates = ['.ts', '.tsx', path.join('index.ts'), path.join('index.tsx')]
    .map((suffix) => normalized.endsWith(suffix) ? normalized : `${normalized}${suffix}`);

  for (const candidate of candidates) {
    if (fs.existsSync(candidate) && fs.statSync(candidate).isFile()) {
      return path.normalize(candidate);
    }
  }

  return null;
}

function resolveImportPath(fromFile, moduleSpecifier) {
  if (!moduleSpecifier.startsWith('.')) {
    return null;
  }

  const basePath = path.resolve(path.dirname(fromFile), moduleSpecifier);
  const direct = normalizeResolvedFile(basePath);
  if (direct) {
    return direct;
  }

  for (const extension of sourceExtensions) {
    const candidate = normalizeResolvedFile(`${basePath}${extension}`);
    if (candidate) {
      return candidate;
    }
  }

  for (const extension of sourceExtensions) {
    const candidate = normalizeResolvedFile(path.join(basePath, `index${extension}`));
    if (candidate) {
      return candidate;
    }
  }

  return null;
}

function collectExports(filePath) {
  const sourceText = fs.readFileSync(filePath, 'utf8');
  const sourceFile = ts.createSourceFile(filePath, sourceText, ts.ScriptTarget.Latest, true, filePath.endsWith('.tsx') ? ts.ScriptKind.TSX : ts.ScriptKind.TS);
  const exports = [];

  for (const statement of sourceFile.statements) {
    if (ts.isExportAssignment(statement)) {
      exports.push({
        name: 'default',
        line: sourceFile.getLineAndCharacterOfPosition(statement.getStart()).line + 1,
        column: sourceFile.getLineAndCharacterOfPosition(statement.getStart()).character + 1,
      });
      continue;
    }

    if (!statement.modifiers?.some((modifier) => modifier.kind === ts.SyntaxKind.ExportKeyword)) {
      continue;
    }

    if (statement.modifiers.some((modifier) => modifier.kind === ts.SyntaxKind.DefaultKeyword)) {
      exports.push({
        name: 'default',
        line: sourceFile.getLineAndCharacterOfPosition(statement.getStart()).line + 1,
        column: sourceFile.getLineAndCharacterOfPosition(statement.getStart()).character + 1,
      });
      continue;
    }

    if (ts.isFunctionDeclaration(statement) || ts.isInterfaceDeclaration(statement) || ts.isTypeAliasDeclaration(statement) || ts.isClassDeclaration(statement) || ts.isEnumDeclaration(statement)) {
      if (statement.name) {
        exports.push({
          name: statement.name.text,
          line: sourceFile.getLineAndCharacterOfPosition(statement.getStart()).line + 1,
          column: sourceFile.getLineAndCharacterOfPosition(statement.getStart()).character + 1,
        });
      }
      continue;
    }

    if (ts.isVariableStatement(statement)) {
      for (const declaration of statement.declarationList.declarations) {
        if (ts.isIdentifier(declaration.name)) {
          exports.push({
            name: declaration.name.text,
            line: sourceFile.getLineAndCharacterOfPosition(declaration.getStart()).line + 1,
            column: sourceFile.getLineAndCharacterOfPosition(declaration.getStart()).character + 1,
          });
        }
      }
      continue;
    }

    if (ts.isExportDeclaration(statement) && statement.exportClause && ts.isNamedExports(statement.exportClause)) {
      for (const element of statement.exportClause.elements) {
        exports.push({
          name: element.name.text,
          line: sourceFile.getLineAndCharacterOfPosition(element.getStart()).line + 1,
          column: sourceFile.getLineAndCharacterOfPosition(element.getStart()).character + 1,
        });
      }
    }
  }

  return exports;
}

function collectImportUsage(allFiles) {
  const usageMap = new Map();

  for (const filePath of allFiles) {
    const sourceText = fs.readFileSync(filePath, 'utf8');
    const sourceFile = ts.createSourceFile(filePath, sourceText, ts.ScriptTarget.Latest, true, filePath.endsWith('.tsx') ? ts.ScriptKind.TSX : ts.ScriptKind.TS);

    for (const statement of sourceFile.statements) {
      if (!ts.isImportDeclaration(statement) || !ts.isStringLiteral(statement.moduleSpecifier)) {
        continue;
      }

      const resolved = resolveImportPath(filePath, statement.moduleSpecifier.text);
      if (!resolved) {
        continue;
      }

      const used = usageMap.get(resolved) ?? new Set();
      const clause = statement.importClause;

      if (clause?.name) {
        used.add('default');
      }

      if (clause?.namedBindings) {
        if (ts.isNamespaceImport(clause.namedBindings)) {
          used.add('*');
        } else {
          for (const element of clause.namedBindings.elements) {
            used.add(element.propertyName?.text ?? element.name.text);
          }
        }
      }

      usageMap.set(resolved, used);
    }
  }

  return usageMap;
}

function buildUnusedExportDiagnostics(targetFiles, allFiles) {
  const usageMap = collectImportUsage(allFiles);

  return targetFiles.map((filePath) => {
    const exported = collectExports(filePath);
    const usage = usageMap.get(filePath) ?? new Set();
    const messages = exported
      .filter(({ name }) => !usage.has('*') && !usage.has(name))
      .map(({ name, line, column }) => ({
        ruleId: 'local/no-unused-public-export',
        severity: 2,
        message: `exported declaration '${name}' is not imported from another module`,
        line,
        column,
      }));

    return {
      filePath,
      messages,
      suppressedMessages: [],
      errorCount: messages.length,
      fatalErrorCount: 0,
      warningCount: 0,
      fixableErrorCount: 0,
      fixableWarningCount: 0,
      usedDeprecatedRules: [],
    };
  });
}

function mergeResults(eslintResults, customResults, targetFiles) {
  const resultMap = new Map();

  for (const result of [...eslintResults, ...customResults]) {
    const existing = resultMap.get(result.filePath);
    if (!existing) {
      resultMap.set(result.filePath, {
        filePath: result.filePath,
        messages: [...result.messages],
        suppressedMessages: result.suppressedMessages ?? [],
        errorCount: result.errorCount ?? 0,
        fatalErrorCount: result.fatalErrorCount ?? 0,
        warningCount: result.warningCount ?? 0,
        fixableErrorCount: result.fixableErrorCount ?? 0,
        fixableWarningCount: result.fixableWarningCount ?? 0,
        usedDeprecatedRules: result.usedDeprecatedRules ?? [],
      });
      continue;
    }

    existing.messages.push(...result.messages);
    existing.errorCount += result.errorCount ?? 0;
    existing.fatalErrorCount += result.fatalErrorCount ?? 0;
    existing.warningCount += result.warningCount ?? 0;
    existing.fixableErrorCount += result.fixableErrorCount ?? 0;
    existing.fixableWarningCount += result.fixableWarningCount ?? 0;
  }

  return targetFiles.map((filePath) => resultMap.get(filePath) ?? {
    filePath,
    messages: [],
    suppressedMessages: [],
    errorCount: 0,
    fatalErrorCount: 0,
    warningCount: 0,
    fixableErrorCount: 0,
    fixableWarningCount: 0,
    usedDeprecatedRules: [],
  });
}

const targetFiles = collectTargetFiles(process.argv.slice(2));
if (targetFiles.length === 0) {
  console.log('[]');
  process.exit(0);
}

const eslint = new ESLint({
  cwd: rootDir,
  overrideConfigFile: path.join(rootDir, 'eslint.config.js'),
});
const eslintResults = (await eslint.lintFiles(targetFiles))
  .map((result) => ({
    ...result,
    messages: result.messages.filter((message) => message.severity === 2),
    suppressedMessages: [],
    warningCount: 0,
    fixableWarningCount: 0,
    errorCount: result.messages.filter((message) => message.severity === 2).length,
  }));
const customResults = buildUnusedExportDiagnostics(targetFiles, collectSourceFiles());
const mergedResults = mergeResults(eslintResults, customResults, targetFiles);

console.log(JSON.stringify(mergedResults));

const hasErrors = mergedResults.some((result) => result.errorCount > 0 || result.fatalErrorCount > 0);
process.exit(hasErrors ? 1 : 0);
