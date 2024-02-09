// organization.go

package main

import (
    "context"
    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "net/http"
)

type Organization struct {
    ID                  string              `json:"organization_id" bson:"_id,omitempty"`
    Name                string              `json:"name"`
    Description         string              `json:"description"`
    OrganizationMembers []OrganizationMember `json:"organization_members"`
}

type OrganizationMember struct {
    Name        string `json:"name"`
    Email       string `json:"email"`
    AccessLevel string `json:"access_level"`
}

func CreateOrganization(c *gin.Context) {
    var org Organization
    if err := c.BindJSON(&org); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Insert the organization into MongoDB
    _, err := organizationCollection.InsertOne(context.Background(), org)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create organization"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"organization_id": org.ID})
}

func ReadOrganization(c *gin.Context) {
    organizationID := c.Param("organization_id")

    var org Organization
    err := organizationCollection.FindOne(context.Background(), bson.M{"_id": organizationID}).Decode(&org)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
        return
    }

    c.JSON(http.StatusOK, org)
}

func ReadAllOrganizations(c *gin.Context) {
    cursor, err := organizationCollection.Find(context.Background(), bson.M{})
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch organizations"})
        return
    }
    defer cursor.Close(context.Background())

    var organizations []Organization
    if err := cursor.All(context.Background(), &organizations); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode organizations"})
        return
    }

    c.JSON(http.StatusOK, organizations)
}

func UpdateOrganization(c *gin.Context) {
    organizationID := c.Param("organization_id")

    var updatedOrg Organization
    if err := c.BindJSON(&updatedOrg); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    _, err := organizationCollection.UpdateOne(context.Background(), bson.M{"_id": organizationID}, bson.M{"$set": updatedOrg})
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update organization"})
        return
    }

    c.JSON(http.StatusOK, updatedOrg)
}

func DeleteOrganization(c *gin.Context) {
    organizationID := c.Param("organization_id")

    _, err := organizationCollection.DeleteOne(context.Background(), bson.M{"_id": organizationID})
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete organization"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "organization deleted successfully"})
}

func InviteUserToOrganization(c *gin.Context) {
    organizationID := c.Param("organization_id")

    // Retrieve the organization from MongoDB
    var org Organization
    err := organizationCollection.FindOne(context.Background(), bson.M{"_id": organizationID}).Decode(&org)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
        return
    }

    // Check if the requesting user has the necessary permissions to invite users
    // (e.g., only admins or certain roles can invite users)
    // This check should be implemented based on your application's specific requirements

    var inviteData struct {
        UserEmail string `json:"user_email"`
    }
    if err := c.BindJSON(&inviteData); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Check if the user with the provided email exists in the database
    var invitedUser User
    err = userCollection.FindOne(context.Background(), bson.M{"email": inviteData.UserEmail}).Decode(&invitedUser)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
        return
    }

    // Add the invited user to the organization members
    org.OrganizationMembers = append(org.OrganizationMembers, OrganizationMember{
        Name:        invitedUser.Name,
        Email:       invitedUser.Email,
        AccessLevel: "member", // Assign access level based on your application's logic
    })

    // Update the organization in MongoDB
    _, err = organizationCollection.UpdateOne(context.Background(), bson.M{"_id": organizationID}, bson.M{"$set": org})
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to invite user"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "user invited to organization successfully"})
}
