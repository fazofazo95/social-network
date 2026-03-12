const FormContainer = ({ children, onSubmit, className = "" }) => {
  return (
    <div className="w-full max-w-md m-auto mt-25 mb-20 rounded-xl overflow-hidden shadow-custom border border-purple-500/30">
      <form 
        className={`bg-[#1a1a2e] px-6 pt-7 pb-8 ${className}`}
        onSubmit={onSubmit}
      >
        {children}
      </form>
    </div>
  );
};

export default FormContainer;
