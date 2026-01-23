const FormContainer = ({ children, onSubmit, className = "" }) => {
  return (
    <div className="w-full max-w-md m-auto mt-25 rounded-lg overflow-hidden shadow-custom">
      <form 
        className={`bg-(--color-customPurple) px-6 pt-7 pb-8 mb-4 ${className}`}
        onSubmit={onSubmit}
      >
        {children}
      </form>
    </div>
  );
};

export default FormContainer;
