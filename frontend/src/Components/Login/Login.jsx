import { useState } from "react"
import { useNavigate } from "react-router-dom";
import UserApi from "../../api/UserApi"

export default function Login() {
    const [values, setValues] = useState({});
    const navigate = useNavigate();
    const [showPassword, setShowPassword] = useState(false);
    const handleChange = (e) => {
        setValues({...values, [e.target.name]: e.target.value});
    }

    const handleSubmit = async (e) => {
        e.preventDefault();

        const form = e.target;
        const formData = new FormData(form);

        const api = UserApi;

        await api.LogIn(formData.get("email"), formData.get("password"))
        
        navigate("/main")
        // API request класс на ооп в отдельном файле с запросами с инкапсуляцией сам url в .env
        // fetch! или axios и в логин тож самое
        // TODO passwords do not match

    }


    return <>
                <h2 className="text-[#518592] font-bold text-2xl">Login</h2>
                <p className="text-[#518592] text-sm mt-4">If you already a member, log in</p>
                <form onSubmit={handleSubmit} className="flex flex-col gap-4" >
                    <input onChange={handleChange} className="p-2 mt-8 rounded-xl border placeholder:opacity-50 " type="email" placeholder="email@mail.com" name="email" required />

                    <div className="relative">
                        <input onChange={handleChange} className="p-2 rounded-xl border w-full placeholder:opacity-50 "
                         type={showPassword ? "text" : "password"} placeholder="password" name="password" required />
                        <button 
                        type="button"
                        className="absolute top-1/2 right-6 -translate-y-1/2"
                        onClick={() => setShowPassword(!showPassword)}>
                            {showPassword ? 
                                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="gray" viewBox="0 0 16 16">
                            <path d="M13.359 11.238C15.06 9.72 16 8 16 8s-3-5.5-8-5.5a7 7 0 0 0-2.79.588l.77.771A6 6 0 0 1 8 3.5c2.12 0 3.879 1.168 5.168 2.457A13 13 0 0 1 14.828 8q-.086.13-.195.288c-.335.48-.83 1.12-1.465 1.755q-.247.248-.517.486z"/>
                            <path d="M11.297 9.176a3.5 3.5 0 0 0-4.474-4.474l.823.823a2.5 2.5 0 0 1 2.829 2.829zm-2.943 1.299.822.822a3.5 3.5 0 0 1-4.474-4.474l.823.823a2.5 2.5 0 0 0 2.829 2.829"/>
                            <path d="M3.35 5.47q-.27.24-.518.487A13 13 0 0 0 1.172 8l.195.288c.335.48.83 1.12 1.465 1.755C4.121 11.332 5.881 12.5 8 12.5c.716 0 1.39-.133 2.02-.36l.77.772A7 7 0 0 1 8 13.5C3 13.5 0 8 0 8s.939-1.721 2.641-3.238l.708.709zm10.296 8.884-12-12 .708-.708 12 12z"/>
                            </svg>
                            :
                            <svg xmlns="http://www.w3.org/2000/svg" 
                            width="16" height="16" fill="gray" viewBox="0 0 16 16">
                            <path d="M16 8s-3-5.5-8-5.5S0 8 0 8s3 5.5 8 5.5S16 8 16 8M1.173 8a13 13 0 0 1 1.66-2.043C4.12 4.668 5.88 3.5 8 3.5s3.879 1.168 5.168 2.457A13 13 0 0 1 14.828 8q-.086.13-.195.288c-.335.48-.83 1.12-1.465 1.755C11.879 11.332 10.119 12.5 8 12.5s-3.879-1.168-5.168-2.457A13 13 0 0 1 1.172 8z"/>
                            <path d="M8 5.5a2.5 2.5 0 1 0 0 5 2.5 2.5 0 0 0 0-5M4.5 8a3.5 3.5 0 1 1 7 0 3.5 3.5 0 0 1-7 0"/>
                            </svg>
                            }
                        </button>
                    </div>

                    <button className="bg-[#518592] rounded-xl text-white py-2 hover:scale-105 duration-300">Log in</button>
                </form>

                <div className="mt-10 grid grid-cols-3 items-center text-gray-500 ">
                    <hr className="border-gray-500" />
                    <p className="text-center">OR</p>
                    <hr className="border-gray-500" />
                </div>

                <div>
                    <p className="mt-5 text-xs py-4">Forgot your password?</p>
                </div>

            </>
}

