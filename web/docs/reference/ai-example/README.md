# AI example reference

Static reference copied from the shadcn AI example set. It is **not** compiled,
linted, or bundled: it lives outside `tsconfig.app.json`'s `src` include and
outside the Vite root, so its imports (`@/components/...`, `ai`, `@tanstack/ai`)
are never resolved.

Use it to compare layout and composition against the production code:

| Reference file      | Production counterpart                                |
| ------------------- | ----------------------------------------------------- |
| `chat.tsx`          | `src/features/ai-ops/components/chat-page.tsx`        |
| `message.tsx`       | `src/features/ai-ops/components/chat-message.tsx`     |
| `prompt-input.tsx`  | `src/features/ai-ops/components/chat-composer.tsx`    |
| `plan.tsx`          | `src/features/ai-ops/components/chat-trace.tsx`       |
| `confirmation.tsx`  | `src/features/ai-ops/components/chat-approval-card.tsx` |
| `reasonning.tsx`    | `src/features/ai-ops/components/chat-message.tsx` (`ThinkingBlock`) |

The production chat is built on `useChat` from `@tanstack/ai-react` with a
ConnectRPC → AG-UI transport (`src/features/ai-ops/api/ai-ops-connection.ts`);
these examples target a server-side TanStack AI endpoint, so do not copy the
data layer from here.
