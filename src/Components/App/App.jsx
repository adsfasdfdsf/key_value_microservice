import { Routes, Route } from "react-router-dom";
import Greeting from "../../Greeting";
import Home from "../../Pages/Identification/Identification";
import Main from "../../Pages/Main/Main";


export default function App() {
    return <>
        <Routes>
             <Route path="/" element={ <Greeting /> } />
             <Route path="/login" element={ <Home showRegistrageionWindow={false} /> } />
             <Route path="/signup" element={ <Home showRegistrageionWindow={true} /> } />
            <Route path="/main" element={ <Main /> } />
        </Routes>
    </>
}