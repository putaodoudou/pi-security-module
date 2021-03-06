package signingapi

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/function61/pi-security-module/accountevent"
	"github.com/function61/pi-security-module/signingapi/signingapitypes"
	"github.com/function61/pi-security-module/state"
	"github.com/function61/pi-security-module/util"
	"github.com/function61/pi-security-module/util/eventbase"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"net/http"
)

func lookupSignerByPubKey(pubKeyMarshaled []byte) (ssh.Signer, *state.InsecureAccount, error) {
	for _, account := range state.Inst.State.Accounts {
		for _, secret := range account.Secrets {
			if secret.SshPrivateKey == "" {
				continue
			}

			signer, err := ssh.ParsePrivateKey([]byte(secret.SshPrivateKey))
			if err != nil {
				panic(err)
			}

			publicKey := signer.PublicKey()

			// TODO: is there better way to compare than marshal result?
			if bytes.Equal(pubKeyMarshaled, publicKey.Marshal()) {
				return signer, &account, nil
			}
		}
	}

	return nil, nil, errors.New("privkey not found by pubkey")
}

func expectedAuthHeader() string {
	bearerHash := sha256.Sum256([]byte(state.Inst.GetMasterPassword() + ":sshagent"))

	return "Bearer " + hex.EncodeToString(bearerHash[:])
}

func Setup(router *mux.Router) {
	log.Printf("signingapi expected auth: %s", expectedAuthHeader())

	router.HandleFunc("/_api/signer/publickeys", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if util.ErrorIfSealed(w, r, state.Inst.IsUnsealed()) {
			return
		}

		if r.Header.Get("Authorization") != expectedAuthHeader() {
			util.CommandCustomError(w, r, "invalid_auth_header", errors.New("Authorization failed"), http.StatusForbidden)
			return
		}

		resp := signingapitypes.NewPublicKeysResponse()

		for _, account := range state.Inst.State.Accounts {
			for _, secret := range account.Secrets {
				if secret.SshPrivateKey == "" {
					continue
				}

				signer, err := ssh.ParsePrivateKey([]byte(secret.SshPrivateKey))
				if err != nil {
					panic(err)
				}

				publicKey := signer.PublicKey()

				pitem := signingapitypes.PublicKeyResponseItem{
					Format:  publicKey.Type(),
					Blob:    publicKey.Marshal(),
					Comment: account.Title,
				}

				resp.PublicKeys = append(resp.PublicKeys, pitem)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))

	router.HandleFunc("/_api/signer/sign", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if util.ErrorIfSealed(w, r, state.Inst.IsUnsealed()) {
			return
		}

		if r.Header.Get("Authorization") != expectedAuthHeader() {
			util.CommandCustomError(w, r, "invalid_auth_header", errors.New("Authorization failed"), http.StatusForbidden)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			util.CommandCustomError(w, r, "unable_to_read_body", err, http.StatusBadRequest)
			return
		}

		var input signingapitypes.SignRequestInput
		if err := json.Unmarshal(body, &input); err != nil {
			util.CommandCustomError(w, r, "unable_to_parse_json", err, http.StatusBadRequest)
			return
		}

		signer, account, err := lookupSignerByPubKey(input.PublicKey)
		if err != nil {
			util.CommandCustomError(w, r, "privkey_for_pubkey_not_found", err, http.StatusBadRequest)
			return
		}

		signature, err := signer.Sign(rand.Reader, input.Data)
		if err != nil {
			util.CommandCustomError(w, r, "signing_failed", err, http.StatusInternalServerError)
			return
		}

		state.Inst.EventLog.Append(accountevent.SecretUsed{
			Event:   eventbase.NewEvent(),
			Account: account.Id,
			Type:    accountevent.SecretUsedTypeSshSigning,
		})

		resp := signingapitypes.SignResponse{
			Signature: signature,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
}
