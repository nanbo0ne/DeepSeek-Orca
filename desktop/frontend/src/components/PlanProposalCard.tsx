import { Markdown } from "./Markdown";

export function PlanProposalCard({ plan }: { plan: string }) {
  const text = plan.trim();
  if (!text) return null;
  return (
    <article className="plan-proposal-card">
      <div className="plan-proposal-card__body">
        <Markdown text={text} />
      </div>
    </article>
  );
}
