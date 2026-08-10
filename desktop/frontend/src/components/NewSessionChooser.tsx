import { useEffect, useMemo, useRef, useState } from "react";
import { ArrowUpRight, Folder, FolderOpen, Search, Sparkles, X } from "lucide-react";
import { app } from "../lib/bridge";
import type { ProjectNode } from "../lib/types";
import { useT } from "../lib/i18n";
import { useMountTransition } from "../lib/useMountTransition";

type NewSessionScope = "global" | "project";

interface ProjectChoice {
  key: string;
  label: string;
  root: string;
}

function collectProjects(nodes: ProjectNode[]): ProjectChoice[] {
  const out: ProjectChoice[] = [];
  const seen = new Set<string>();
  const visit = (node: ProjectNode) => {
    if (node.kind === "project" && node.root && !seen.has(node.root)) {
      seen.add(node.root);
      out.push({ key: node.key || node.root, label: node.label || node.root, root: node.root });
    }
    for (const child of node.children ?? []) visit(child);
  };
  for (const node of nodes) visit(node);
  return out;
}

function rootTail(root: string): string {
  const cleaned = root.replace(/[\\/]+$/, "");
  const parts = cleaned.split(/[\\/]/).filter(Boolean);
  return parts[parts.length - 1] || root;
}

export function NewSessionChooser({
  open,
  onClose,
  onChoose,
  onPickProjectFolder,
}: {
  open: boolean;
  onClose: () => void;
  onChoose: (scope: NewSessionScope, workspaceRoot: string) => Promise<void> | void;
  onPickProjectFolder: () => Promise<void> | void;
}) {
  const t = useT();
  const { mounted, status } = useMountTransition(open, 180);
  const [projects, setProjects] = useState<ProjectChoice[]>([]);
  const [query, setQuery] = useState("");
  const [busyKey, setBusyKey] = useState<string | null>(null);
  const searchRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (!open) return;
    let cancelled = false;
    setQuery("");
    setBusyKey(null);
    requestAnimationFrame(() => searchRef.current?.focus());
    void app.ListProjectTree()
      .then((nodes) => {
        if (!cancelled) setProjects(collectProjects(Array.isArray(nodes) ? nodes : []));
      })
      .catch(() => {
        if (!cancelled) setProjects([]);
      });
    return () => {
      cancelled = true;
    };
  }, [open]);

  useEffect(() => {
    if (!mounted) return;
    const onKey = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        event.preventDefault();
        onClose();
      }
    };
    document.addEventListener("keydown", onKey);
    return () => document.removeEventListener("keydown", onKey);
  }, [mounted, onClose]);

  const filteredProjects = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return projects;
    return projects.filter((project) => `${project.label}\n${project.root}`.toLowerCase().includes(q));
  }, [projects, query]);

  if (!mounted) return null;

  const choose = async (scope: NewSessionScope, workspaceRoot: string, key: string) => {
    if (busyKey !== null) return;
    setBusyKey(key);
    try {
      await onChoose(scope, workspaceRoot);
    } finally {
      setBusyKey(null);
    }
  };

  const pickProjectFolder = async () => {
    if (busyKey !== null) return;
    setBusyKey("pick-project");
    try {
      await onPickProjectFolder();
      onClose();
    } finally {
      setBusyKey(null);
    }
  };

  return (
    <div className="new-session-backdrop" data-state={status} onMouseDown={onClose}>
      <section className="new-session" data-state={status} role="dialog" aria-modal="true" aria-labelledby="new-session-title" onMouseDown={(event) => event.stopPropagation()}>
        <button className="new-session__close" type="button" aria-label={t("common.close")} onClick={onClose}>
          <X size={16} />
        </button>

        <div className="new-session__hero">
          <div className="new-session__mark" aria-hidden="true">
            <Sparkles size={20} />
          </div>
          <h2 id="new-session-title">{t("newSession.title")}</h2>
        </div>

        <button
          type="button"
          className="new-session__primary"
          disabled={busyKey !== null}
          onClick={() => void choose("global", "", "global")}
        >
          <span className="new-session__icon"><Folder size={17} /></span>
          <span className="new-session__body">
            <strong>{t("newSession.independent")}</strong>
            <span>{t("newSession.independentHint")}</span>
          </span>
          <ArrowUpRight size={16} />
        </button>

        <div className="new-session__search">
          <Search size={15} />
          <input
            ref={searchRef}
            value={query}
            onChange={(event) => setQuery(event.currentTarget.value)}
            placeholder={t("newSession.searchPlaceholder")}
          />
        </div>

        <div className="new-session__section-title">{t("newSession.projects")}</div>
        <div className="new-session__projects">
          {filteredProjects.length > 0 ? (
            filteredProjects.map((project) => (
              <button
                key={project.key}
                type="button"
                className="new-session__project"
                disabled={busyKey !== null}
                onClick={() => void choose("project", project.root, project.key)}
              >
                <span className="new-session__icon new-session__icon--project"><FolderOpen size={17} /></span>
                <span className="new-session__body">
                  <strong>{project.label || rootTail(project.root)}</strong>
                  <span>{project.root}</span>
                </span>
              </button>
            ))
          ) : (
            <div className="new-session__empty">{projects.length === 0 ? t("newSession.noProjects") : t("newSession.noProjectMatches")}</div>
          )}
        </div>

        <button type="button" className="new-session__folder" disabled={busyKey !== null} onClick={() => void pickProjectFolder()}>
          <FolderOpen size={16} />
          <span>{t("newSession.pickProjectFolder")}</span>
        </button>
      </section>
    </div>
  );
}
