import DataRow from "../../UI/DataRow/DataRow";

function isSubsequence(sub, str) {
  let subIdx = 0;
  
  for (let char of str) {
    if (char === sub[subIdx]) {
      subIdx++;
    }
    if (subIdx === sub.length) return true;
  }
  
  return subIdx === sub.length;
}

export default function TableKeyValue({showAll, search, values, showFiles}) {

    return <>
        
        <div className="min-w-3xl bg-gray-100 flex flex-col gap-4 rounded-2xl p-4">
            <div className="flex flex-row justify-between items-center w-full gap-4">
            <div className="block bg-gray-300 w-full p-1 text-center 
            rounded-xl text-xl text-bold">Key</div>
            <div>→</div>
            <div className="block bg-gray-300 w-full p-1 text-center 
            rounded-xl text-xl">Value</div>
            </div>
            <hr></hr>

            {
            values.map((el) => {
                return isSubsequence(search, el.key) && <DataRow showAll={showAll} key={el.key} valueKey={el.key} value={el.value} />
            }) 
            }


        </div>
    </>
}