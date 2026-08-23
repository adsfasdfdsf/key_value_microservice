import { useState, useEffect } from "react"
import Header from "../../Components/Header/Header"
import Table from "../../Components/Table/Table"
import ActionBar from "../../Components/ActionBar/ActionBar"
import ModalAddValue from "../../Components/ModalAddValue/ModalAddValue";
import UserApi from "../../api/UserApi";


export default function Main() {
    const api = UserApi;

    const [search, setSearch] = useState("");
    const [isOpen, setIsOpen] = useState(false);
    const [values, setValues] = useState([])
    const [showAll, setShowAll] = useState(false)

    const showModal = () => {
        setIsOpen(true);
    }

    const closeModal = () => {
        setIsOpen(false);
    }
    
    const loadValues = async () => {
            try {
                const data = await UserApi.GetUserValues();
                setValues(data.data);
            } catch (error) {
                console.error(error);
            }
        }

    useEffect(() => {
        loadValues();
    }, []);

    return <>
        <section className="bg-gray-50 min-h-screen flex flex-col items-center justify-start bg-">
            {isOpen && (<ModalAddValue closeModal={closeModal} onValueAdded={loadValues} />)}
            
            <Header setSearch={setSearch}  />
            <Table showAll={showAll} search={search} values={values ?? []} />
            
            {/* Table:container with search filter and toggle files or data
            + table with data of files 
            libraries for dnd */}

            <ActionBar showAll={showAll} setShowAll={setShowAll} showModal={showModal} />
        </section>
    </>
}