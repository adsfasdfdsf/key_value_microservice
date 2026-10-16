import { useState } from "react";



export default function ActionBar({showAll, setShowAll, showModal}){

    return <>
        <div className="fixed bottom-2 items-center flex flex-row gap-4 justify-center bg-white/0 
        backdrop-blur-xl m-4 rounded-lg w-[95%]">
            <button 
            onClick={() => {showModal()}}
            className="bg-[#518592] rounded-xl text-white px-4
            py-2 hover:scale-105 duration-300">Add Value</button>
            <button 
            onClick={() => {setShowAll(!showAll)}}
            className="bg-[#518592] rounded-xl text-white px-4
            py-2 hover:scale-105 duration-300">{showAll ? "Hide All" : "Show All"}</button>
        </div>
    </>
}