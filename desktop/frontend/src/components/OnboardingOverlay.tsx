import { useEffect, useMemo, useRef, useState, type KeyboardEvent } from "react";
import { ArrowRight, Check, Download, KeyRound, Monitor, SkipForward } from "lucide-react";
import logoSymbol from "../assets/logo-symbol.png";
import { app } from "../lib/bridge";
import { normalizeLocalAICatalog } from "../lib/localAI";
import { DEFAULT_ONBOARDING_ROUTE, type OnboardingRoute } from "../lib/onboarding";
import type { LocalAICatalogView, OnboardingState } from "../lib/types";

export function OnboardingOverlay({ onComplete }: { onComplete: () => void }) {
  const [state, setState] = useState<OnboardingState | null>(null);
  const [route, setRoute] = useState<OnboardingRoute>(DEFAULT_ONBOARDING_ROUTE);
  const [providerID, setProviderID] = useState("deepseek");
  const [key, setKey] = useState("");
  const [modelID, setModelID] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [localCatalog, setLocalCatalog] = useState<LocalAICatalogView | null>(null);
  const dialogRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    void app.GetOnboardingState()
      .then((next) => {
        setState(next as OnboardingState);
        const deepseek = next.providers.find((p) => p.name === "deepseek");
        setProviderID(deepseek?.name || next.providers.find((p) => p.added)?.name || next.providers[0]?.name || "deepseek");
      })
      .catch((e) => setError(`无法读取模型配置：${String(e)}`));
  }, []);
  useEffect(() => {
    const frame = requestAnimationFrame(() => {
      dialogRef.current?.querySelector<HTMLElement>("input, select, button:not([disabled])")?.focus();
    });
    return () => cancelAnimationFrame(frame);
  }, [route]);
  useEffect(() => {
    if (route !== "local") return;
    setError(null);
    setLocalCatalog(null);
    void app.GetLocalAICatalog()
      .then((next) => setLocalCatalog(normalizeLocalAICatalog(next)))
      .catch((e) => setError(String(e)));
  }, [route]);

  const selected = useMemo(() => state?.providers.find((p) => p.name === providerID), [providerID, state]);
  const changeRoute = (next: OnboardingRoute) => { setError(null); setRoute(next); };
  const keepFocusInDialog = (event: KeyboardEvent<HTMLDivElement>) => {
    if (event.key !== "Tab") return;
    const focusable = Array.from(dialogRef.current?.querySelectorAll<HTMLElement>(
      'button:not([disabled]), input:not([disabled]), select:not([disabled]), [tabindex]:not([tabindex="-1"])',
    ) || []).filter((element) => element.offsetParent !== null);
    if (focusable.length === 0) { event.preventDefault(); dialogRef.current?.focus(); return; }
    const first = focusable[0];
    const last = focusable[focusable.length - 1];
    if (event.shiftKey && document.activeElement === first) { event.preventDefault(); last.focus(); }
    else if (!event.shiftKey && document.activeElement === last) { event.preventDefault(); first.focus(); }
  };
  const complete = async () => { setBusy(true); setError(null); try { await app.CompleteOnboarding(); onComplete(); } catch (e) { setError(String(e)); } finally { setBusy(false); } };
  const connect = async () => {
    if (!providerID || !key.trim()) { setError("请选择供应商并填写 API Key"); return; }
    setBusy(true); setError(null);
    try { await app.ConnectProviderPreset(providerID, key.trim()); await complete(); } catch (e) { setError(String(e)); setBusy(false); }
  };
  const installLocal = async () => {
    if (!localCatalog?.supported) { setError("本地模型首版仅支持 Windows"); return; }
    setBusy(true); setError(null);
    try { await app.StartLocalRuntimeInstall(localCatalog.hardware.recommendedRuntime); const model = modelID || localCatalog.hardware.recommendedModel || localCatalog.models[0]?.id; if (model) await app.StartLocalModelDownload(model); await complete(); } catch (e) { setError(String(e)); setBusy(false); }
  };

  return <div className="onboarding">
    <div ref={dialogRef} className="onboarding__card" role="dialog" aria-modal="true" aria-labelledby="onboarding-title" aria-describedby="onboarding-description" tabIndex={-1} onKeyDown={keepFocusInDialog}>
      <img src={logoSymbol} className="onboarding__logo" alt="O.R.C.A." draggable={false} />
      <div id="onboarding-title" className="onboarding__title">O.R.C.A. for Windows</div>
      <div id="onboarding-description" className="onboarding__tag">Open Reasoning &amp; Computing Agent · 先连接模型，再开始工作</div>
      {route === "deepseek" && <>
        <div className="onboarding__privacy">输入 DeepSeek API Key 即可开始。密钥仅保存到本机凭据配置；也可以跳过，稍后连接其他供应商或本地模型。</div>
        <div className="onboarding__provider-heading"><KeyRound size={18} /><span><strong>DeepSeek 官方 API</strong><small>https://api.deepseek.com</small></span></div>
        <label className="onboarding__label" htmlFor="onboarding-deepseek-key">DEEPSEEK_API_KEY</label>
        <input id="onboarding-deepseek-key" className="onboarding__input" type="password" autoComplete="off" value={key} onChange={(e) => setKey(e.target.value)} placeholder="sk-..." />
        {error && <div className="onboarding__error" role="alert">{error}</div>}
        <button className="onboarding__submit" disabled={busy || !key.trim()} onClick={() => { setProviderID("deepseek"); void connect(); }}>{busy ? "正在验证…" : "验证并开始使用"}</button>
        <div className="onboarding__choices onboarding__choices--secondary">
          <button className="onboarding__choice" onClick={() => changeRoute("providers")}><KeyRound size={18} /><span><strong>其他供应商</strong><small>OpenAI、Anthropic、国内模型服务等</small></span><ArrowRight size={16} /></button>
          <button className="onboarding__choice" onClick={() => changeRoute("local")}><Monitor size={18} /><span><strong>安装本地 AI</strong><small>检测硬件并推荐 llama.cpp 与 Qwen</small></span><ArrowRight size={16} /></button>
        </div>
        <button className="onboarding__skip" disabled={busy} onClick={() => void complete()}><SkipForward size={14} />暂不配置，进入应用</button>
      </>}
      {route === "providers" && <>
        <label className="onboarding__label" htmlFor="onboarding-provider">供应商</label>
        <select id="onboarding-provider" className="onboarding__input" value={providerID} onChange={(e) => setProviderID(e.target.value)}>{(state?.providers || []).map((p) => <option key={p.name} value={p.name}>{p.name} · {p.baseUrl}</option>)}</select>
        <label className="onboarding__label" htmlFor="onboarding-key">{selected?.apiKeyEnv || "API Key"}</label>
        <input id="onboarding-key" className="onboarding__input" type="password" autoComplete="off" value={key} onChange={(e) => setKey(e.target.value)} placeholder="只保存在本机安全配置中" />
        {error && <div className="onboarding__error" role="alert">{error}</div>}
        <button className="onboarding__submit" disabled={busy} onClick={() => void connect()}>{busy ? "正在验证…" : "验证并继续"}</button>
        <button className="onboarding__skip" disabled={busy} onClick={() => { setProviderID("deepseek"); changeRoute("deepseek"); }}>返回 DeepSeek 快速设置</button>
      </>}
      {route === "local" && <>
        <div className="onboarding__privacy">本地运行器只监听 127.0.0.1。16GB 级显卡优先推荐 Qwen3.8-27B IQ3_XXS，并会为当前空闲显存自动降低上下文或 GPU 层数。</div>
        {!localCatalog && !error && <div className="onboarding__privacy">正在读取本机硬件与模型目录…</div>}
        {localCatalog && <select className="onboarding__input" value={modelID || localCatalog.hardware.recommendedModel || ""} onChange={(e) => setModelID(e.target.value)}>{localCatalog.models.map((model) => <option key={model.id} value={model.id}>{model.name}</option>)}</select>}
        {error && <div className="onboarding__error" role="alert">{error}</div>}
        <button className="onboarding__submit" disabled={busy || !localCatalog?.supported} onClick={() => void installLocal()}>{busy ? "正在准备下载…" : <><Download size={14} />安装并继续</>}</button>
        <button className="onboarding__skip" disabled={busy} onClick={() => changeRoute("deepseek")}>返回 DeepSeek 快速设置</button>
      </>}
      <div className="onboarding__links"><span>升级用户会保留现有会话、供应商和模型配置。</span><span><Check size={13} />可随时在设置中更改</span></div>
    </div>
  </div>;
}
