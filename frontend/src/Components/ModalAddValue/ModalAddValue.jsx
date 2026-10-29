import { useState } from "react";
import UserApi from "../../api/UserApi";

export default function ModalAddValue({ closeModal, onValueAdded }) {
    const [inputValue, setInputValue] = useState({ key: "", value: "" });
    const [error, setError] = useState("");
    const [pending, setPending] = useState(false);

    const changeInput = (event) => {
        setInputValue({ ...inputValue, [event.target.name]: event.target.value });
    };

    const addValue = async (event) => {
        event.preventDefault();
        setError("");
        setPending(true);
        try {
            await UserApi.AddKey(inputValue.key, inputValue.value);
            await onValueAdded();
            closeModal();
        } catch (failure) {
            setError(failure.message);
        } finally {
            setPending(false);
        }
    };

    return <div className="fixed inset-0 z-50 flex items-center justify-center px-4">
        <div onClick={() => !pending && closeModal()} className="absolute inset-0 bg-black/40 backdrop-blur-sm" />
        <div role="dialog" aria-modal="true" aria-labelledby="add-title" className="relative w-full flex flex-col gap-4 max-w-md bg-gray-100 p-8 rounded-xl">
            <h2 id="add-title" className="text-center text-[#518592] font-bold text-2xl">Add new key-value</h2>
            <form onSubmit={addValue} className="flex flex-col gap-4">
                <input onChange={changeInput} type="text" placeholder="key" className="p-2 rounded-xl border w-full placeholder:opacity-50" required name="key" />
                <input onChange={changeInput} type="text" placeholder="value" className="p-2 rounded-xl border w-full placeholder:opacity-50" name="value" />
                {error && <p role="alert" className="text-red-700">{error}</p>}
                <button disabled={pending} className="py-2 bg-[#518592] text-white rounded-xl hover:shadow-xl duration-300 disabled:opacity-50">{pending ? "Saving…" : "Add"}</button>
                <button type="button" disabled={pending} className="bg-gray-200 rounded-xl hover:shadow-md" onClick={closeModal}>Cancel</button>
            </form>
        </div>
    </div>;
}
