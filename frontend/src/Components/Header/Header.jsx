
export default function Header({setSearch}){
    return <>
        <div className="sticky top-0 items-center flex flex-row justify-center bg-[#518592]/25 
        backdrop-blur-xl m-4 rounded-lg w-[95%]">
            <input type="text" onChange={(e) => {setSearch(e.target.value)}}
            className=" m-2 rounded-xl p-2 border-black border-2 outline-0" />
        </div>
    </>
}