import { useState } from "react";
import Login from "../Login/Login";
import SignIn from "../SignIn/SignUp";

export default function Home() {
  const [showRegister, setShowRegister] = useState(false);
  const [isAnimating, setIsAnimating] = useState(false);
  const [moveImage, setMoveImage] = useState(false);



  const changeForm = (e) => {
    setIsAnimating(true);
    setMoveImage(prev => !prev);

    setTimeout(() => {
      setShowRegister(prev => !prev);
      setIsAnimating(false);
    }, 500);
  };

  return (
    <section className="bg-gray-50 min-h-screen flex items-center justify-center">
      <div className="bg-gray-100 flex flex-col sm:flex-row rounded-2xl shadow-lg max-w-3xl p-5 items-center overflow-hidden
      ">

        <div
          className={`sm:w-1/2 px-12 ${
            isAnimating ? "animate-fade-out" : "animate-fade-in"
          } ${showRegister ? `sm:translate-x-full` : `sm:translate-x-0`}`}
        >
          {showRegister ? <SignIn /> : <Login />}
        </div>

        <div className={`hidden sm:block w-1/2 transition-all duration-700 ${moveImage ? "-translate-x-full" : "translate-x-0"}`}>
          <img
            className="rounded-2xl"
            src="/image.png"
            alt=""
          />

          <button onClick={changeForm} className={`
          absolute top-1/2 -translate-y-1/2 right-1/2 translate-x-1/2
            rounded-xl
            px-5 py-2
            bg-white/10
            backdrop-blur-xl
            border border-white/20
            shadow-[inset_0_1px_1px_rgba(255,255,255,0.2)]
            shadow-black/20
            text-white
            hover:bg-white/15
            transition-all duration-300
          ${
            isAnimating ? "animate-fade-out" : "animate-fade-in"
          }`}>
            {`${showRegister ? "Log in" : "Sign in"}`}
          </button>
        </div>


            <button onClick={changeForm} className={`sm:hidden mx-12 bg-[#518592] w-full rounded-xl text-white py-2 hover:scale-105
            transition-all duration-300 
          ${
            isAnimating ? "animate-fade-out" : "animate-fade-in"
          }`}>
            {`${showRegister ? "Log in" : "Sign in"}`}
          </button>


      </div>
    </section>
  );
}