import ReactMarkdown from "react-markdown";

export function MarkdownContent({ source }: { source: string }) {
  return <ReactMarkdown>{source}</ReactMarkdown>;
}
