package user

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/utils"
)

// An end point to create new users
func Create(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	user_collection := "users"

	var user User

	err := utils.ParseJsonFromRequest(request, &user)
	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, response)
		return
	}
	if !utils.IsValidEmail(user.Email) {
		utils.GetError(errors.New("email address is not valid"), http.StatusBadRequest, response)
		return
	}

	// confirm if user_email exists
	result, _ := utils.GetMongoDbDoc(user_collection, bson.M{"email": user.Email})
	if result != nil {
		fmt.Printf("users with email %s exists!", user.Email)
		utils.GetError(errors.New("operation failed"), http.StatusBadRequest, response)
		return
	}

	user.CreatedAt = time.Now()

	detail, _ := utils.StructToMap(user)

	res, err := utils.CreateMongoDbDoc(user_collection, detail)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, response)
		return
	}

	utils.GetSuccess("user created", res, response)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	userId := params["user_id"]

	delete, err := utils.DeleteOneMongoDoc("users", userId)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}
	if delete.DeletedCount == 0 {
		utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("User Deleted Succesfully", nil, w)
}

// helper functions perform CRUD operations on user
func FindUserByID(response http.ResponseWriter, request *http.Request) {
	// Find a user by user ID
	response.Header().Set("content-type", "application/json")

	collectionName := "users"
	userID := mux.Vars(request)["id"]
	objID, err := primitive.ObjectIDFromHex(userID)

	if err != nil {
		utils.GetError(err, http.StatusBadRequest, response)
		return
	}

	res, err := utils.GetMongoDbDoc(collectionName, bson.M{"_id": objID})
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, response)
		return
	}
	utils.GetSuccess("User retrieved successfully", res, response)

}

func UpdateUser(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	// Validate the user ID
	userID := mux.Vars(request)["id"]
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		utils.GetError(errors.New("invalid user ID"), http.StatusBadRequest, response)
		return
	}

	collectionName := "users"
	userExist, err := utils.GetMongoDbDoc(collectionName, bson.M{"_id": objID})
	if err != nil {
		utils.GetError(errors.New("User does not exist"), http.StatusNotFound, response)
		return
	}
	if userExist != nil {
		var user UserUpdate
		if err := utils.ParseJsonFromRequest(request, &user); err != nil {
			utils.GetError(errors.New("bad update data"), http.StatusUnprocessableEntity, response)
			return
		}

		userMap, err := utils.StructToMap(user)
		if err != nil {
			utils.GetError(err, http.StatusInternalServerError, response)
		}

		updateFields := make(map[string]interface{})
		for key, value := range userMap {
			if value != "" {
				updateFields[key] = value
			}
		}
		if len(updateFields) == 0 {
			utils.GetError(errors.New("empty/invalid user input data"), http.StatusBadRequest, response)
			return
		} else {
			updateRes, err := utils.UpdateOneMongoDbDoc(collectionName, userID, updateFields)
			if err != nil {
				utils.GetError(errors.New("user update failed"), http.StatusInternalServerError, response)
				return
			}
			utils.GetSuccess("user successfully updated", updateRes, response)
		}

	}

}
