package actions

import "github.com/forestvpn/cli/auth"

// Login is a method for logging in a user on the Firebase.
// Accepts the deviceID (coming from local file) which indicates wether the device was created on previous login.
// If the deviceID is empty, then should create a new device on login.
//
// See https://firebase.google.com/docs/reference/rest/auth#section-sign-in-email-password for more information.
func (w AuthClientWrapper) Login(email string, password string) error {
	var user_id string
	signinform := auth.SignInForm{}
	emailfield, err := auth.GetEmailField(email)

	if err != nil {
		return err
	}

	signinform.EmailField = emailfield
	user_id = w.AccountsMap.GetUserID(emailfield.Value)

	if len(user_id) == 0 {
		validate := false
		passwordfield, err := auth.GetPasswordField([]byte(password), validate)

		if err != nil {
			return err
		}

		signinform.PasswordField = passwordfield
		response, err := w.AuthClient.SignIn(signinform)

		if err != nil {
			return err
		}

		err = auth.HandleFirebaseSignInResponse(response)

		if err != nil {
			return err
		}

		user_id, err = w.SetUpProfile(response)

		if err != nil {
			return err
		}

		err = w.AccountsMap.AddAccount(signinform.EmailField.Value, user_id)

		if err != nil {
			return err
		}
	}

	active := auth.IsActiveProfile(user_id)

	if !active {
		return auth.SetActiveProfile(user_id)
	}
	return nil
}
