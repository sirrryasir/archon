import React, { useState, useEffect } from 'react';
import { useFPSThrottler } from '../ui/hooks/useFPSThrottler';
import { render, Box, Text, useApp, useInput, Newline } from 'ink';
import TextInput from 'ink-text-input';
import chalk from 'chalk';
import { gatherProjectContext } from '../mcp/index.ts';
import { chatStream, type Message } from '../ai/index.ts';
import { findCommand, parseSlashCommand, getCommands } from '../commands/index.ts';
import type { ChatContext } from '../commands/types.ts';
import {
  createSessionId,
  saveSessionMeta,
  appendMessage as persistMessage,
  generateTitle,
  type SessionMeta,
} from '../session/index.ts';
import { detectMode, type ProjectMode } from '../modes/detect.ts';
import { GREENFIELD_SYSTEM_PROMPT, GREENFIELD_WELCOME } from '../modes/greenfield.ts';
import { buildBrownfieldPrompt, BROWNFIELD_WELCOME } from '../modes/brownfield.ts';
import type { DesignDocs } from '../design/generator.ts';
import { MarkdownRenderer } from '../rendering/markdown.tsx';
import { PlanPreview, type PlanItem } from '../ui/components/PlanPreview.tsx';

chalk.level = 3;

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Banner Component
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

function Banner({ fileCount, model, mode }: { fileCount: number; model: string; mode: ProjectMode }) {
  const modeLabel = mode === 'greenfield' ? 'Greenfield' : 'Brownfield';
  const modeColor = mode === 'greenfield' ? '#10b981' : '#f59e0b';
  const borderColor = mode === 'greenfield' ? '#059669' : '#d97706';

  return (
    <Box flexDirection="column" alignItems="center" marginBottom={1} marginTop={1}>
      <Box borderStyle="round" borderColor={borderColor} paddingX={2} flexDirection="column" alignItems="center">
        <Text bold color="cyan">
          {'    █████╗ ██████╗  ██████╗██╗  ██╗ ██████╗ ███╗   ██╗    '}
          <Newline/>
          {'   ██╔══██╗██╔══██╗██╔════╝██║  ██║██╔═══██╗████╗  ██║    '}
          <Newline/>
          {'   ███████║██████╔╝██║     ███████║██║   ██║██╔██╗ ██║    '}
          <Newline/>
          {'   ██╔══██║██╔══██╗██║     ██╔══██║██║   ██║██║╚██╗██║    '}
          <Newline/>
          {'   ██║  ██║██║  ██║╚██████╗██║  ██║╚██████╔╝██║ ╚████║    '}
          <Newline/>
          {'   ╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═══╝    '}
        </Text>
        <Box marginTop={1}>
          <Text bold color="white">SOFTWARE ARCHITECT</Text>
        </Box>
      </Box>
      
      <Box marginTop={1} gap={2}>
        <Box paddingX={1} borderStyle="single" borderColor={modeColor}>
          <Text color={modeColor}>
            {modeLabel.toUpperCase()}
          </Text>
        </Box>
        <Box paddingX={1} borderStyle="single" borderColor="gray">
          <Text color="gray">
            MODEL: <Text color="white">{model.toUpperCase()}</Text>
          </Text>
        </Box>
        {mode === 'brownfield' && (
          <Box paddingX={1} borderStyle="single" borderColor="gray">
            <Text color="gray">
              FILES: <Text color="white">{fileCount}</Text>
            </Text>
          </Box>
        )}
      </Box>

      {mode === 'greenfield' && (
        <Box marginTop={1}>
          <Text color="gray" italic>System initialized. Ready for discovery phase.</Text>
        </Box>
      )}
    </Box>
  );
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Status Bar Component
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

function StatusBar({ model, messageCount, sessionId }: {
  model: string;
  messageCount: number;
  sessionId: string;
}) {
  return (
    <Box 
      marginTop={1} 
      flexDirection="row" 
      justifyContent="space-between" 
      width="100%" 
      paddingX={1}
      borderStyle="single"
      borderColor="gray"
    >
      <Box gap={2}>
        <Text color="gray">
          <Text color="white" bold>EXIT</Text> ^C
        </Text>
        <Text color="#4b5563">│</Text>
        <Text color="gray">
          <Text color="white" bold>HELP</Text> /?
        </Text>
      </Box>
      
      <Box gap={2}>
        <Text color="gray">
          MSG: <Text color="white">{messageCount}</Text>
        </Text>
        <Text color="#4b5563">│</Text>
        <Text color="gray">
          MODEL: <Text color="#3b82f6">{model.toUpperCase()}</Text>
        </Text>
        <Text color="#4b5563">│</Text>
        <Text color="gray">
          SESSION: <Text color="white">{sessionId.slice(0, 8)}</Text>
        </Text>
      </Box>
    </Box>
  );
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Command Result Display
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

function CommandResult({ output }: { output: string }) {
  return (
    <Box 
      flexDirection="column" 
      marginBottom={1} 
      paddingLeft={2} 
      borderStyle="single" 
      borderLeft 
      borderRight={false} 
      borderTop={false} 
      borderBottom={false} 
      borderColor="#475569"
    >
      <Text color="#cbd5e1">{output}</Text>
    </Box>
  );
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Main Chat Application
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

interface DisplayItem {
  type: 'user' | 'assistant' | 'command-result' | 'plan';
  content: string;
  planItems?: PlanItem[];
}

function ChatApp() {
  const { exit } = useApp();
  const [messages, setMessages] = useState<Message[]>([]);
  const [displayItems, setDisplayItems] = useState<DisplayItem[]>([]);
  const [input, setInput] = useState('');
  const [isProcessing, setIsProcessing] = useState(false);
  const [streamingResponse, setStreamingResponse] = useFPSThrottler('', 20);
  const [isInitializing, setIsInitializing] = useState(true);
  const [fileCount, setFileCount] = useState(0);
  const [model, setModel] = useState(process.env.ARCHON_MODEL ?? 'qwen3:14b');
  const [sessionId] = useState(() => createSessionId());

  // ── Global Keybindings ──
  useInput((input, key) => {
    // Ctrl+C: Exit
    if (key.ctrl && input === 'c') {
      exit();
      return;
    }

    // Ctrl+L: Clear screen
    if (key.ctrl && input === 'l') {
      console.clear();
      return;
    }

    // Esc: Clear input or cancel processing
    if (key.escape) {
      if (isProcessing) {
        // Handle cancel logic if needed
      }
      setInput('');
      return;
    }

    // Alt+M (Meta+M): Toggle Mode
    if (key.meta && input === 'm') {
      const nextMode = mode === 'greenfield' ? 'brownfield' : 'greenfield';
      setMode(nextMode);
      return;
    }
  });
  const [sessionMeta, setSessionMeta] = useState<SessionMeta | null>(null);
  const [mode, setMode] = useState<ProjectMode>('brownfield');
  const [designDocs, setDesignDocs] = useState<DesignDocs | null>(null);

  // Ctrl+C handler
  useInput((input: string, key: any) => {
    if (key.ctrl && input === 'c') {
      exit();
    }
  });

  // Initialize: detect mode, scan workspace, set up system prompt
  useEffect(() => {
    let active = true;
    async function init() {
      const cwd = process.cwd();

      // Step 1: Detect mode
      const detection = await detectMode(cwd);
      if (!active) return;
      setMode(detection.mode);

      let systemContent: string;

      if (detection.mode === 'greenfield') {
        // Greenfield: no project context, use discovery prompt
        systemContent = GREENFIELD_SYSTEM_PROMPT;
        setFileCount(0);

        // Show greenfield welcome message
        setDisplayItems([{
          type: 'command-result',
          content: GREENFIELD_WELCOME,
        }]);
      } else {
        // Brownfield: scan project and inject context
        const projectContext = await gatherProjectContext(cwd);
        const fileLines = projectContext.split('\n').filter((l: string) => l.startsWith('### File:'));
        if (!active) return;
        setFileCount(fileLines.length);

        systemContent = buildBrownfieldPrompt(projectContext);

        // Show brownfield welcome message
        setDisplayItems([{
          type: 'command-result',
          content: BROWNFIELD_WELCOME,
        }]);
      }

      const systemMessage: Message = {
        role: 'system',
        content: systemContent,
      };
      setMessages([systemMessage]);

      // Create session metadata
      const meta: SessionMeta = {
        id: sessionId,
        projectPath: cwd,
        createdAt: Date.now(),
        updatedAt: Date.now(),
        title: 'New Session',
        messageCount: 0,
        model,
      };
      setSessionMeta(meta);
      await saveSessionMeta(meta);

      setIsInitializing(false);
    }
    init();
    return () => { active = false; };
  }, []);

  // Build chat context for commands
  const chatContext: ChatContext = {
    cwd: process.cwd(),
    messages,
    setMessages: (newMsgs: Message[]) => {
      setMessages(newMsgs);
    },
    model,
    setModel: (newModel: string) => {
      setModel(newModel);
    },
    exit,
    projectPath: process.cwd(),
    mode,
    setMode: (newMode: ProjectMode) => {
      setMode(newMode);
    },
    designDocs,
    setDesignDocs: (docs: DesignDocs) => {
      setDesignDocs(docs);
    },
  };

  const handleSubmit = async (query: string) => {
    const trimmed = query.trim();
    if (!trimmed) return;

    setInput('');

    // ── Slash Command Handling ──
    const parsed = parseSlashCommand(trimmed);
    if (parsed) {
      const cmd = findCommand(parsed.name);
      if (cmd) {
        setIsProcessing(true);
        try {
          const result = await cmd.execute(parsed.args, chatContext);
          if (result) {
            setDisplayItems(prev => [
              ...prev,
              { type: 'command-result', content: result },
            ]);
          }
          // Special case: /review adds a user message that needs AI response
          if (cmd.name === 'review') {
            // The review command already added the message, now stream the response
            const currentMessages = [...messages];
            // Get the latest messages (review command may have updated them)
            const latestMessages = chatContext.messages;
            await streamAIResponse(latestMessages);
          }
        } catch (e: any) {
          setDisplayItems(prev => [
            ...prev,
            { type: 'command-result', content: chalk.red(`Error: ${e.message}`) },
          ]);
        } finally {
          setIsProcessing(false);
        }
        return;
      } else {
        // Unknown command
        setDisplayItems(prev => [
          ...prev,
          {
            type: 'command-result',
            content: chalk.yellow(`  Unknown command: /${parsed.name}. Type /help for available commands.`),
          },
        ]);
        return;
      }
    }

    // ── Regular Chat Message ──
    setIsProcessing(true);
    setStreamingResponse('');

    const userMessage: Message = { role: 'user', content: trimmed };
    const newMessages: Message[] = [...messages, userMessage];
    setMessages(newMessages);
    setDisplayItems(prev => [...prev, { type: 'user', content: trimmed }]);

    // Persist message
    await persistMessage(sessionId, userMessage);

    await streamAIResponse(newMessages);
  };

  async function streamAIResponse(messagesToSend: Message[]) {
    try {
      let fullResponse = '';
      for await (const chunk of chatStream(messagesToSend)) {
        fullResponse += chunk;
        setStreamingResponse(fullResponse);
      }

      const assistantMessage: Message = { role: 'assistant', content: fullResponse };
      setMessages(prev => [...prev, assistantMessage]);
      setDisplayItems(prev => [...prev, { type: 'assistant', content: fullResponse }]);

      // Persist
      await persistMessage(sessionId, assistantMessage);

      // Update session meta
      if (sessionMeta) {
        const updatedMeta: SessionMeta = {
          ...sessionMeta,
          updatedAt: Date.now(),
          messageCount: sessionMeta.messageCount + 2,
          title: generateTitle(messages),
          model,
        };
        setSessionMeta(updatedMeta);
        await saveSessionMeta(updatedMeta);
      }
    } catch (e: any) {
      setDisplayItems(prev => [
        ...prev,
        { type: 'assistant', content: `[Error: ${e.message}]` },
      ]);
    } finally {
      setIsProcessing(false);
      setStreamingResponse('');
    }
  }

  const chatMessageCount = displayItems.filter(
    d => d.type === 'user' || d.type === 'assistant'
  ).length;

  const hasContent = displayItems.length > 0;

  return (
    <Box flexDirection="column" paddingX={2} paddingY={1} width="100%">
      {/* Banner — only shown when no messages yet */}
      {!isInitializing && <Banner fileCount={fileCount} model={model} mode={mode} />}

      {/* Messages History */}
      {hasContent && (
        <Box flexDirection="column" marginBottom={1}>
          {displayItems.map((item, idx) => (
            <Box key={idx} flexDirection="column" marginBottom={1}>
              {item.type === 'user' && (
                <>
                  <Box marginBottom={0}>
                    <Text bold color="#10b981">YOU</Text>
                  </Box>
                  <Box paddingLeft={2}>
                    <Text color="white">{item.content}</Text>
                  </Box>
                </>
              )}
              {item.type === 'assistant' && (
                <>
                  <Box marginBottom={0}>
                    <Text bold color="cyan">ARCHON</Text>
                  </Box>
                  <Box paddingLeft={2}>
                    <MarkdownRenderer text={item.content} />
                  </Box>
                </>
              )}
              {item.type === 'command-result' && (
                <CommandResult output={item.content} />
              )}
              {item.type === 'plan' && item.planItems && (
                <PlanPreview items={item.planItems} />
              )}
            </Box>
          ))}

          {/* Streaming response */}
          {isProcessing && streamingResponse && (
            <Box flexDirection="column" marginBottom={1}>
              <Box marginBottom={0}>
                <Text bold color="cyan">ARCHON</Text>
              </Box>
              <Box paddingLeft={2}>
                <MarkdownRenderer text={streamingResponse} />
              </Box>
            </Box>
          )}
        </Box>
      )}

      {/* The Input Box */}
      <Box
        flexDirection="column"
        borderStyle="round"
        borderColor={isProcessing ? "gray" : "cyan"}
        paddingX={2}
        paddingY={1}
        width="100%"
      >
        <Box flexDirection="row" alignItems="center">
          <Box marginRight={1}>
            <Text bold color={isProcessing ? "gray" : "cyan"}>❯</Text>
          </Box>

          {isInitializing ? (
            <Text color="gray">{mode === 'greenfield' ? 'Preparing architect mode...' : 'Scanning workspace and loading project context...'}</Text>
          ) : isProcessing ? (
            <Box gap={1}>
              <Text color="yellow" bold italic>Archon is thinking...</Text>
              <Text color="gray">(Press Esc to cancel)</Text>
            </Box>
          ) : (
            <TextInput
              value={input}
              onChange={setInput}
              onSubmit={handleSubmit}
              placeholder='Ask anything... or type /help for commands'
            />
          )}
        </Box>

        {/* Quick action hints */}
        {!isInitializing && !isProcessing && !hasContent && (
          <Box marginTop={1} flexDirection="row" gap={2}>
            <Text color="#06b6d4">✧ Architecture</Text>
            <Text color="#64748b">│</Text>
            <Text color="#3b82f6">✧ Dependencies</Text>
            <Text color="#64748b">│</Text>
            <Text color="#8b5cf6">✧ Security</Text>
          </Box>
        )}
      </Box>

      {/* Status Bar */}
      {!isInitializing && (
        <Box flexDirection="column">
          <StatusBar 
            model={model} 
            messageCount={displayItems.filter(i => i.type !== 'command-result').length} 
            sessionId={sessionId} 
          />
        </Box>
      )}
    </Box>
  );
}

export async function startChatSession() {
  console.clear();
  const app = render(<ChatApp />);
  await app.waitUntilExit();
}
