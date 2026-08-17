import { useEffect, useMemo, useState } from "react";
import { ArrowRight, Check, Download, KeyRound, Monitor, SkipForward } from "lucide-react";
import logo from "../assets/logo.png";
import { app } from "../lib/bridge";
import type { OnboardingState } from "../lib/types";

type Route = "choose" | "cloud" | "local";

export function OnboardingOverlay({ onComplete }: { onComplete: () => void }) {
  const [state, setState] = useState<OnboardingState | null>(null);
  const [route, setRoute] = useState<Route>("choose");
  const [providerID, setProviderID] = useState("");
  const [key, setKey] = useState("");
  const [modelID, setModelID] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [localCatalog, setLocalCatalog] = useState<any>(null);

  useEffect(() => {
    void app.GetOnboardingState().then((next) => { setState(next as OnboardingState); setProviderID(next.providers.find((p) => p.added)?.name || next.providers[0]?.name || ""); });
  }, []);
  useEffect(() => { if (route === "local") void app.GetLocalAICatalog().then(setLocalCatalog).catch((e) => setError(String(e))); }, [route]);

  const selected = useMemo(() => state?.providers.find((p) => p.name === providerID), [providerID, state]);
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
    <div className="onboarding__card">
      <img src={logo} className="onboarding__logo" alt="O.R.C.A" draggable={false} />
      <div className="onboarding__title">O.R.C.A for Windows</div>
      <div className="onboarding__tag">Open Reasoning &amp; Computing Agent · 先连接模型，再开始工作</div>
      {route === "choose" && <>
        <div className="onboarding__privacy">没有 DeepSeek 依赖。你可以连接任意支持的云端模型，或在 Windows 上安装可选的本地 AI；截图与本地文件遵循当前权限设置。</div>
        <div className="onboarding__choices">
          <button className="onboarding__choice" onClick={() => setRoute("cloud")}><KeyRound size={20} /><span><strong>连接云端模型</strong><small>选择供应商并验证密钥</small></span><ArrowRight size={16} /></button>
          <button className="onboarding__choice" onClick={() => setRoute("local")}><Monitor size={20} /><span><strong>安装本地 AI</strong><small>检测硬件并推荐 llama.cpp 与 Qwen</small></span><ArrowRight size={16} /></button>
        </div>
        <button className="onboarding__skip" disabled={busy} onClick={() => void complete()}><SkipForward size={14} />稍后设置</button>
      </>}
      {route === "cloud" && <>
        <label className="onboarding__label" htmlFor="onboarding-provider">供应商</label>
        <select id="onboarding-provider" className="onboarding__input" value={providerID} onChange={(e) => setProviderID(e.target.value)}>{(state?.providers || []).map((p) => <option key={p.name} value={p.name}>{p.name} · {p.baseUrl}</option>)}</select>
        <label className="onboarding__label" htmlFor="onboarding-key">{selected?.apiKeyEnv || "API Key"}</label>
        <input id="onboarding-key" className="onboarding__input" type="password" autoComplete="off" value={key} onChange={(e) => setKey(e.target.value)} placeholder="只保存在本机安全配置中" />
        {error && <div className="onboarding__error" role="alert">{error}</div>}
        <button className="onboarding__submit" disabled={busy} onClick={() => void connect()}>{busy ? "正在验证…" : "验证并继续"}</button>
        <button className="onboarding__skip" disabled={busy} onClick={() => setRoute("choose")}>返回</button>
      </>}
      {route === "local" && <>
        <div className="onboarding__privacy">本地运行器只监听 127.0.0.1。16GB 级显卡优先推荐 Qwen3.8-27B IQ3_XXS，并会为当前空闲显存自动降低上下文或 GPU 层数。</div>
        {localCatalog && <select className="onboarding__input" value={modelID || localCatalog.hardware.recommendedModel || ""} onChange={(e) => setModelID(e.target.value)}>{localCatalog.models.map((m: any) => <option key={m.id} value={m.id}>{m.name}</option>)}</select>}
        {error && <div className="onboarding__error" role="alert">{error}</div>}
        <button className="onboarding__submit" disabled={busy || !localCatalog?.supported} onClick={() => void installLocal()}>{busy ? "正在准备下载…" : <><Download size={14} />安装并继续</>}</button>
        <button className="onboarding__skip" disabled={busy} onClick={() => setRoute("choose")}>返回</button>
      </>}
      {error && route === "choose" && <div className="onboarding__error" role="alert">{error}</div>}
      <div className="onboarding__links"><span>升级用户会保留现有会话、供应商和模型配置。</span><span><Check size={13} />可随时在设置中更改</span></div>
    </div>
  </div>;
}
