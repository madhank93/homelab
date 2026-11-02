// Rewrites `[label](src:path/in/repo)` into a link to that file at the ref the
// docs describe, so a release cut changes one constant instead of every page.
import { visit } from 'unist-util-visit';

export default function remarkSrcLinks({ repo, ref }) {
  return (tree) => {
    visit(tree, 'link', (node) => {
      if (node.url.startsWith('src:')) {
        node.url = `${repo}/blob/${ref}/${node.url.slice(4)}`;
      }
    });
  };
}
