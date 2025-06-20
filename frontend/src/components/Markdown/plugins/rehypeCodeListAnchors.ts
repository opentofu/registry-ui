import { visit } from 'unist-util-visit';
import { Element, Node, Parent } from 'hast';

/**
 * A rehype plugin that adds IDs to list items that start with inline code
 * and wraps the code content in a link to itself.
 * This is particularly useful for schema documentation where attributes are listed.
 */
export function rehypeCodeListAnchors() {
  return (tree: Node) => {
    visit(tree, 'element', (node: Element, index: number, parent: Parent) => {
      // Check if this is a list item
      if (node.tagName === 'li' && node.children && node.children.length > 0) {
        let codeElement: Element | null = null;
        let codeParent: Element = node;
        let codeIndex = 0;
        
        // Check if first child is code directly
        if (node.children[0].type === 'element' && node.children[0].tagName === 'code') {
          codeElement = node.children[0];
          codeParent = node;
          codeIndex = 0;
        }
        // Check if first child is a paragraph containing code
        else if (node.children[0].type === 'element' && 
                node.children[0].tagName === 'p' && 
                node.children[0].children &&
                node.children[0].children[0] &&
                node.children[0].children[0].type === 'element' &&
                node.children[0].children[0].tagName === 'code') {
          codeElement = node.children[0].children[0];
          codeParent = node.children[0];
          codeIndex = 0;
        }
        
        // Process the code element if found
        if (codeElement && 
            codeElement.children &&
            codeElement.children.length > 0 &&
            codeElement.children[0].type === 'text') {
          
          // Extract the code content
          const codeText = codeElement.children[0].value;
          
          // Create a URL-friendly ID from the code text
          // Remove backticks and special characters, convert to lowercase
          const id = codeText
            .replace(/[`'"]/g, '')
            .replace(/[^a-zA-Z0-9_-]/g, '-')
            .replace(/-+/g, '-')
            .replace(/^-|-$/g, '')
            .toLowerCase();
          
          // Add the ID to the list item
          if (id) {
            node.properties = node.properties || {};
            node.properties.id = `attr-${id}`;
            
            // Create an anchor element that wraps the code
            const anchor: Element = {
              type: 'element',
              tagName: 'a',
              properties: {
                href: `#attr-${id}`,
                className: ['hover:underline', 'cursor-pointer']
              },
              children: [codeElement]
            };
            
            // Replace the code element with the anchor
            codeParent.children[codeIndex] = anchor;
          }
        }
      }
    });
  };
}