import { useState } from "react"
import UserApi from "../../api/UserApi";



export default function ModalAddValue({closeModal, onValueAdded}) {
    const [inputValue, setInputValue] = useState({})
    

    const changeInput = (e) => {
        setInputValue({...inputValue, [e.target.name]: e.target.value});
    }
    
    const AddValue = async () => {
        const api = UserApi
        console.log(inputValue.key, inputValue.value)
        await api.AddKey(inputValue.key, inputValue.value)
        onValueAdded();
    }
    
    
    return <>
        <div className="fixed inset-0 z-50 flex items-center justify-center px-4">
            <div 
            onClick={() => closeModal()} 
            className="absolute inset-0 bg-black/40 backdrop-blur-sm"
            />

            <div className="relative w-full flex flex-col gap-4 max-w-md bg-gray-100 p-8 rounded-xl">
                <h2 className="text-center text-[#518592]
                text-bold text-2xl">Add new key-value</h2>
                <form className="flex flex-col gap-4">
                    <input onChange={changeInput} type="text" placeholder="key" className="p-2 rounded-xl border w-full placeholder:opacity-50"
                    required name="key" />
                    <input onChange={changeInput} type="text" placeholder="value" className="p-2 rounded-xl border w-full placeholder:opacity-50" 
                    required name="value" />
                    <button type="button" className="py-2 bg-[#518592] text-white rounded-xl
                    hover:scale-105 hover:shadow-xl duration-300" onClick={AddValue}> Add</button>
                    <button type="button" className="bg-gray-200 rounded-xl
                    hover:shadow-md hover:scale-105 duration-300"
                    onClick={() => closeModal()}>Cancel</button>
                </form>
            </div> 
        </div>
    </>
}