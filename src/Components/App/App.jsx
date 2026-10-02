import { useEffect, useSyncExternalStore } from "react";
import { Routes, Route, Navigate } from "react-router-dom";
// import Greeting from "../../Greeting";
import Home from "../../Pages/Identification/Identification";
import Main from "../../Pages/Main/Main";
import { getSessionStatus, subscribeSession, restoreSession } from "../../api/ApiClient";

export default function App() {
    const status = useSyncExternalStore(subscribeSession, getSessionStatus);
    useEffect(() => { restoreSession(); }, []);

    if (status === "checking") {
        return <div className="min-h-screen flex items-center justify-center">Проверяем сессию…</div>;
    }
    const authenticated = status === "authenticated";
    return <>
        {status === "error" && <div role="alert" className="fixed top-0 inset-x-0 z-50 bg-amber-100 p-3 text-center">
            <p>Не удалось проверить сессию. Можно войти вручную или повторить проверку.</p>
            <button type="button" onClick={restoreSession}>Повторить проверку</button>
        </div>}
        <Routes>
        {/* <Route path="/" element={<Greeting />} /> */}
        <Route path="/" element={<Navigate to={authenticated ? "/main" : "/login"} replace />} />
        <Route path="/login" element={authenticated
            ? <Navigate to="/main" replace />
            : <Home key="login" showRegistrageionWindow={false} />} />
        <Route path="/signup" element={authenticated
            ? <Navigate to="/main" replace />
            : <Home key="signup" showRegistrageionWindow={true} />} />
        <Route path="/main" element={authenticated ? <Main /> : <Navigate to="/login" replace />} />
        <Route path="*" element={<Navigate to={authenticated ? "/main" : "/login"} replace />} />
        </Routes>
    </>;
}
