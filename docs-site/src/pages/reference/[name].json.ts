import fs from 'node:fs';
import path from 'node:path';

const docsRoot = path.resolve(import.meta.dirname, '../../../../docs');

export function getStaticPaths() {
  return fs
    .readdirSync(docsRoot, { withFileTypes: true })
    .filter((entry) => entry.isFile() && entry.name.endsWith('.json'))
    .map((entry) => ({
      params: {
        name: path.parse(entry.name).name,
      },
      props: {
        absolutePath: path.join(docsRoot, entry.name),
      },
    }));
}

export function GET({ props }: { props: { absolutePath: string } }) {
  const body = fs.readFileSync(props.absolutePath, 'utf8');

  return new Response(body, {
    headers: {
      'content-type': 'application/json; charset=utf-8',
    },
  });
}
