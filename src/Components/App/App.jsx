import { Routes, Route } from "react-router-dom";
import Login from "../../Pages/Login/Login";
import Greeting from "../../Greeting";
import SignIn from "../../Pages/SignIn/SignIn";
import Home from "../../Pages/Home/Home";

export default function App() {
    return <>
        <Routes>
             <Route path="/login" element={<Login />} />
             <Route path="/signin" element={<SignIn />} />
             <Route path="/" element={ <Greeting /> } />
             <Route path="/home" element={ <Home /> } />
        </Routes>
    </>
}