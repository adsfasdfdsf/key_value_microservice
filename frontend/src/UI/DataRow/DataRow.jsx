import { useState } from "react";



export default function DataRow(props){
    const [hover, setHover] = useState(false);
    const onHover = () => {
        setHover(true);
    };

    const onLeave = () => {
        setHover(false);
    };
    
    
    return <div className="flex flex-row justify-between items-center w-full gap-4">
        <div className="block bg-gray-300 w-full p-1 text-center 
        rounded-xl text-lg text-bold">{props.valueKey}</div>
        <div>→</div>
        <div className="block bg-gray-300 w-full p-1 text-center 
        rounded-xl text-lg" onMouseEnter={onHover} onMouseLeave={onLeave}>{hover || props.showAll ? props.value : "*".repeat(props.value.length)}</div>
    </div>
}