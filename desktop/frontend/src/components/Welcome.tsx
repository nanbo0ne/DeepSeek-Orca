import { useEffect, useState } from "react";
import logoWordmark from "../assets/logo-wordmark.png";
import { app, onWelcomeSuggestions } from "../lib/bridge";
import { useT } from "../lib/i18n";

// Welcome is the empty-state landing: a one-liner, the input affordances
// (/ commands, @ files, Enter), and a few clickable example prompts that send
// immediately so a first turn is one click away.

export function Welcome({ onPrompt }: { onPrompt: (text: string) => void }) {
  const t = useT();
  const fallbackExamples = [t("welcome.ex1"), t("welcome.ex2"), t("welcome.ex3"), t("welcome.ex4")];
  const [generatedExamples, setGeneratedExamples] = useState<string[]>([]);
  const examples = generatedExamples.length === 4 ? generatedExamples : fallbackExamples;

  useEffect(() => {
    let active = true;
    void app.GetWelcomeSuggestions().then((prompts) => {
      if (active && prompts.length === 4) setGeneratedExamples(prompts);
    }).catch(() => undefined);
    const off = onWelcomeSuggestions((prompts) => {
      if (prompts.length === 4) setGeneratedExamples(prompts);
    });
    return () => {
      active = false;
      off();
    };
  }, []);
  return (
    <div className="welcome welcome--brand">
      <span className="welcome__brand">
        <img src={logoWordmark} className="welcome__brand-logo" alt="DeepSeek-Orca" draggable={false} />
      </span>
      <h2 className="welcome__title">{t("welcome.title")}</h2>
      <div className="welcome__tag">{t("welcome.tagline")}</div>

      <div className="welcome__hints">
        <span>
          <kbd>/</kbd> {t("welcome.hintCommands")}
        </span>
        <span>
          <kbd>@</kbd> {t("welcome.hintFiles")}
        </span>
        <span>
          <kbd>Enter</kbd> {t("welcome.hintSend")}
        </span>
      </div>

      <div className="welcome__examples">
        {examples.map((ex) => (
          <button key={ex} className="welcome__ex" onClick={() => onPrompt(ex)}>
            {ex}
          </button>
        ))}
      </div>
    </div>
  );
}
