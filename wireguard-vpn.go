package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsssm"
	"github.com/aws/jsii-runtime-go"

	"github.com/aws/constructs-go/constructs/v10"

	_ "embed"
)

type WireguardVPNStackProps struct {
	awscdk.StackProps
}

//go:embed user-data.sh
var wireguardInstanceUserData string

func NewWireguardVPNStack(scope constructs.Construct, id string, props *WireguardVPNStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// We only need public subnets to run our VPN within.
	vpc := awsec2.NewVpc(stack, jsii.String("wireguardVpc"), &awsec2.VpcProps{
		NatGateways: jsii.Number(0),
	})

	param := awsssm.StringParameter_FromStringParameterName(stack, jsii.String("wireguardPrivateKey"), jsii.String("wireguardPrivateKey"))

	wireguardSG := awsec2.NewSecurityGroup(stack, jsii.String("wireguardSecurityGroup"), &awsec2.SecurityGroupProps{
		Vpc:              vpc,
		AllowAllOutbound: jsii.Bool(true),
		Description:      jsii.String("Enable Wireguard VPN."),
	})
	wireguardSG.AddIngressRule(awsec2.Peer_AnyIpv4(), awsec2.Port_Udp(jsii.Number(51820)), jsii.String("Allow Wireguard inbound (IP v4)."), jsii.Bool(false))
	wireguardSG.AddIngressRule(awsec2.Peer_AnyIpv6(), awsec2.Port_Udp(jsii.Number(51820)), jsii.String("Allow Wireguard inbound (IP v6)."), jsii.Bool(false))

	wireguardInstance := awsec2.NewInstance(stack, jsii.String("wireguardInstance"), &awsec2.InstanceProps{
		InstanceType:              awsec2.InstanceType_Of(awsec2.InstanceClass_BURSTABLE4_GRAVITON, awsec2.InstanceSize_SMALL),
		MachineImage:              awsec2.MachineImage_FromSsmParameter(jsii.String("/aws/service/canonical/ubuntu/server/20.04/stable/current/arm64/hvm/ebs-gp2/ami-id"), &awsec2.SsmParameterImageOptions{}),
		Vpc:                       vpc,
		AllowAllOutbound:          jsii.Bool(true),
		SecurityGroup:             wireguardSG,
		UserData:                  awsec2.MultipartUserData_Custom(&wireguardInstanceUserData),
		UserDataCausesReplacement: jsii.Bool(true),
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PUBLIC,
		},
	})
	wireguardRole := wireguardInstance.Role()
	wireguardRole.AddManagedPolicy(awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonSSMManagedInstanceCore")))
	param.GrantRead(wireguardInstance)

	// Give the machine a static IP.
	ip := awsec2.NewCfnEIP(stack, jsii.String("wireguardElasticIp"), &awsec2.CfnEIPProps{
		InstanceId: wireguardInstance.InstanceId(),
	})

	awscdk.NewCfnOutput(stack, jsii.String("wireguardInstanceId"), &awscdk.CfnOutputProps{
		Value:      wireguardInstance.InstanceId(),
		ExportName: jsii.String("wireguardInstanceId"),
	})
	awscdk.NewCfnOutput(stack, jsii.String("wireguardPublicIp"), &awscdk.CfnOutputProps{
		Value:      ip.Ref(),
		ExportName: jsii.String("wireguardPublicIp"),
	})

	return stack
}

func main() {
	app := awscdk.NewApp(nil)

	NewWireguardVPNStack(app, "WireguardVPNStack", &WireguardVPNStackProps{
		awscdk.StackProps{},
	})

	app.Synth(nil)
}
