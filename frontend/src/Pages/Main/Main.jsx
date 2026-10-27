import { useState, useEffect } from "react";
import Header from "../../Components/Header/Header";
import TableKeyValue from "../../Components/TableKeyValue/TableKeyValue";
import ActionBar from "../../Components/ActionBar/ActionBar";
import ModalAddValue from "../../Components/ModalAddValue/ModalAddValue";
import UserApi from "../../api/UserApi";

export default function Main() {
    const [search, setSearch] = useState("");
    const [isOpen, setIsOpen] = useState(false);
    const [values, setValues] = useState([]);
    const [showAll, setShowAll] = useState(false);
    const [error, setError] = useState("");

    const loadValues = async () => {
        try {
            const data = await UserApi.GetUserValues();
            setValues(data.data);
            setError("");
        } catch (failure) {
            setError(failure.message);
        }
    };

    useEffect(() => {
        let active = true;
        UserApi.GetUserValues().then(data => {
            if (active) setValues(data.data);
        }).catch(failure => {
            if (active) setError(failure.message);
        });
        return () => { active = false; };
    }, []);

    return <section className="bg-gray-50 min-h-screen flex flex-col items-center justify-start px-4">
        {isOpen && <ModalAddValue closeModal={() => setIsOpen(false)} onValueAdded={loadValues} />}
        <Header setSearch={setSearch} />
        {error && <div role="alert" className="p-4 text-red-700">
            {error} <button type="button" onClick={loadValues}>Повторить</button>
        </div>}
        <TableKeyValue showAll={showAll} search={search} values={values} />
        <ActionBar showAll={showAll} setShowAll={setShowAll} showModal={() => setIsOpen(true)} />
    </section>;
}
