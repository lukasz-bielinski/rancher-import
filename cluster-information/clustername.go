package clustername

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
)

func getClusterNameFromRoleARN(arn string) string {
	// The ARN format is: arn:aws:iam::<account-id>:role/eks.amazonaws.com/v1/role/<role-name>/eks.amazonaws.com/v1/cluster/<cluster-name>
	parts := strings.Split(arn, "/")
	if len(parts) > 2 && parts[len(parts)-2] == "cluster" {
		return parts[len(parts)-1]
	}
	return ""
}

func GetClusterName() (string, error) {
	// Create a new session
	sess := session.Must(session.NewSession())
	svc := ec2metadata.New(sess)

	// Get IAM info
	iamInfo, err := svc.IAMInfo()
	if err != nil {
		return "", err
	}

	return getClusterNameFromRoleARN(iamInfo.InstanceProfileArn), nil
}
