import Image from "next/image";

const SearchBar = ({ placeholder = "Search...", icon = "/search_icon.svg", onChange, className = "" }) => {
  return (
    <div className={`relative ${className}`}>
      <Image
        src={icon}
        alt="Search Icon"
        width={20}
        height={20}
        className="absolute left-4 top-1.75 w-2.5"
      />
      <input
        type="text"
        placeholder={placeholder}
        onChange={onChange}
        className="rounded-md bg-gray-200 ml-2 pl-6 text-black"
      />
    </div>
  );
};

export default SearchBar;
