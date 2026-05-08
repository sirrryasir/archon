import React from 'react';
import { Box, Text } from 'ink';

export interface PlanItem {
  path: string;
  type: 'modify' | 'new' | 'delete';
  description?: string;
}

interface PlanPreviewProps {
  items: PlanItem[];
  title?: string;
}

/**
 * PlanPreview
 * 
 * A professional component to visualize proposed changes in a tree-like structure.
 * Inspired by 'aider' and 'claude-code' plan previews.
 */
export function PlanPreview({ items, title = 'PROPOSED ARCHITECTURAL CHANGES' }: PlanPreviewProps) {
  if (items.length === 0) return null;

  return (
    <Box flexDirection="column" marginY={1} borderStyle="round" borderColor="cyan" paddingX={2}>
      <Box marginBottom={1}>
        <Text bold color="white">{title}</Text>
      </Box>

      {items.map((item, idx) => {
        const color = 
          item.type === 'new' ? '#10b981' : 
          item.type === 'delete' ? '#ef4444' : 
          '#3b82f6';
        
        const typeLabel = item.type.toUpperCase();

        return (
          <Box key={idx} marginBottom={0}>
            <Box width={10}>
              <Text color={color}>[{typeLabel}]</Text>
            </Box>
            <Box flexGrow={1}>
              <Text color="white">{item.path}</Text>
              {item.description && (
                <Text color="gray"> — {item.description}</Text>
              )}
            </Box>
          </Box>
        );
      })}

      <Box marginTop={1}>
        <Text color="gray" italic>Total items: {items.length}</Text>
      </Box>
    </Box>
  );
}
