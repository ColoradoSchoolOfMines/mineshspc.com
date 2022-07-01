import { Formik, Field } from "formik";
import axios from "axios";

import "./InterestForm.scss";

function InterestForm() {
  return (
    <div className="InterestForm">
      <Formik
        initialValues={{ email: "", iama: "Teacher" }}
        validate={(values) => {
          const errors = {};

          if (!values.email) {
            errors.email = "You must enter an email address";
          } else if (
            !/^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}$/i.test(values.email)
          ) {
            errors.email = "Invalid email address";
          }

          return errors;
        }}
        onSubmit={async (values, { setSubmitting }) => {
          try {
            await axios.put("http://localhost:8090/register_interest", values, {
              headers: {
                "Content-Type": "application/json",
              },
            });
            console.log("got here");
            alert("did a thing");
          } catch (error) {
            console.log(error);
          } finally {
            setSubmitting(false);
          }
        }}
      >
        {({
          values,
          errors,
          touched,
          handleChange,
          handleBlur,
          handleSubmit,
          isSubmitting,
          /* and other goodies */
        }) => {
          const anyErrors = Object.values(errors).reduce(
            (prev, cur) => prev || cur !== "",
            false
          );

          return (
            <form
              onSubmit={handleSubmit}
              className={
                (touched.email && anyErrors) || (touched.email && !errors.email)
                  ? "was-validated"
                  : "needs-validation"
              }
              noValidate
            >
              <div className="form-group form-row">
                <div className="col-sm-12 form-description form-text text-center">
                  Enter your information below if you are interested in the 2023
                  competition.
                </div>
              </div>

              <div className="form-row">
                <div className="col-sm-1"></div>
                <label className="col-sm-1 col-form-label">I am a</label>
                <div className="col-sm-3">
                  <div
                    className="btn-group btn-group-toggle"
                    data-toggle="buttons"
                    role="group"
                    aria-labelledby="i-am-a"
                  >
                    {["Student", "Teacher"].map((option, i) => (
                      <label
                        key={i}
                        className={`btn btn-${
                          values.iama === option ? "primary" : "secondary"
                        } ${values.iama === option ? "active" : ""}`}
                      >
                        <Field type="radio" name="iama" value={option} />
                        {option}
                      </label>
                    ))}
                  </div>
                </div>
                <div className="col-sm-5">
                  <input
                    id="email"
                    type="email"
                    name="email"
                    className="form-control"
                    aria-describedby="email-help"
                    required
                    placeholder="Email"
                    onChange={handleChange}
                    onBlur={handleBlur}
                    value={values.email}
                  />
                  {errors.email && touched.email && (
                    <div className="invalid-feedback">{errors.email}</div>
                  )}
                  <small id="email-help" className="form-text text-muted">
                    We will send you an email when it's time to register.
                  </small>
                </div>
                <div className="col-sm-1">
                  <button
                    type="submit"
                    className={`btn ${
                      !isSubmitting && touched.email && !anyErrors
                        ? "btn-primary"
                        : "btn-secondary"
                    }`}
                    disabled={isSubmitting}
                  >
                    Submit
                  </button>
                </div>
                <div className="col-sm-1"></div>
              </div>
            </form>
          );
        }}
      </Formik>
    </div>
  );
}

export default InterestForm;
