package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/J0sh0909/rift/internal"
	"github.com/spf13/cobra"
)

// ---------------------------------------------------------------------------
// AWS flags
// ---------------------------------------------------------------------------

var (
	awsAllFlag        bool
	awsHardFlag       bool
	awsYesFlag        bool
	awsAMIFlag        string
	awsNameFlag       string
	awsTypeFlag       string
	awsRegionFlag     string
	awsUserFlag       string
	awsEncryptAllFlag bool
)

// ---------------------------------------------------------------------------
// Lazy AWS backend
// ---------------------------------------------------------------------------

var awsBackend *internal.AWSBackend

func requireAWS() {
	if awsBackend != nil {
		return
	}
	// Load settings for AWS_REGION (best-effort; .env is optional for AWS).
	s, _ := internal.LoadSettings()
	region := awsRegionFlag
	if region == "" {
		region = s.AWSRegion // may still be empty — SDK falls back to ~/.aws/config
	}
	var err error
	awsBackend, err = internal.NewAWSBackend(region)
	if err != nil {
		internal.LogError(internal.ErrAWS, "", "initializing AWS: %s", err)
		os.Exit(1)
	}
}

// ---------------------------------------------------------------------------
// rift aws
// ---------------------------------------------------------------------------

var awsCmd = &cobra.Command{
	Use:   "aws",
	Short: "Manage AWS EC2 instances",
}

// ---------------------------------------------------------------------------
// rift aws list
// ---------------------------------------------------------------------------

var awsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List EC2 instances",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		requireAWS()
		instances, err := awsBackend.ListInstances(awsAllFlag)
		if err != nil {
			internal.LogError(internal.ErrAWS, "", "listing instances: %s", err)
			os.Exit(1)
		}
		if len(instances) == 0 {
			fmt.Println("No instances found.")
			return
		}
		fmt.Printf("%-20s %-20s %-12s %-12s %-16s %-16s %s\n",
			"INSTANCE ID", "NAME", "STATE", "TYPE", "PUBLIC IP", "PRIVATE IP", "LAUNCHED")
		fmt.Println(strings.Repeat("-", 110))
		for _, i := range instances {
			launched := ""
			if !i.LaunchTime.IsZero() {
				launched = i.LaunchTime.Format("2006-01-02 15:04")
			}
			pub := i.PublicIP
			if pub == "" {
				pub = "-"
			}
			name := i.Name
			if name == "" {
				name = "-"
			}
			fmt.Printf("%-20s %-20s %-12s %-12s %-16s %-16s %s\n",
				i.ID, name, i.State, i.Type, pub, i.PrivateIP, launched)
		}
	},
}

// ---------------------------------------------------------------------------
// rift aws start
// ---------------------------------------------------------------------------

var awsStartCmd = &cobra.Command{
	Use:   "start <instance-id> [instance-id...]",
	Short: "Start stopped EC2 instances",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		requireAWS()
		if err := awsBackend.StartInstances(args); err != nil {
			internal.LogError(internal.ErrAWSStartFailed, "", "%s", err)
			os.Exit(1)
		}
		for _, id := range args {
			fmt.Printf("%s → starting...\n", id)
		}
		for _, id := range args {
			if err := awsBackend.WaitUntilRunning(id, 5*time.Minute); err != nil {
				internal.LogError(internal.ErrAWSStartFailed, id, "waiting for running: %s", err)
				continue
			}
			inst, err := awsBackend.GetInstance(id)
			if err != nil {
				internal.LogError(internal.ErrAWSNotFound, id, "%s", err)
				continue
			}
			pub := inst.PublicIP
			if pub == "" {
				pub = "(no public IP)"
			}
			fmt.Printf("%s → running — %s\n", id, pub)
		}
	},
}

// ---------------------------------------------------------------------------
// rift aws stop
// ---------------------------------------------------------------------------

var awsStopCmd = &cobra.Command{
	Use:   "stop <instance-id> [instance-id...]",
	Short: "Stop running EC2 instances",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		requireAWS()
		if err := awsBackend.StopInstances(args, awsHardFlag); err != nil {
			internal.LogError(internal.ErrAWSStopFailed, "", "%s", err)
			os.Exit(1)
		}
		for _, id := range args {
			fmt.Printf("%s → stopping...\n", id)
		}
		for _, id := range args {
			if err := awsBackend.WaitUntilStopped(id, 5*time.Minute); err != nil {
				internal.LogError(internal.ErrAWSStopFailed, id, "waiting for stopped: %s", err)
				continue
			}
			fmt.Printf("%s → stopped\n", id)
		}
	},
}

// ---------------------------------------------------------------------------
// rift aws create
// ---------------------------------------------------------------------------

var awsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Launch a new EC2 instance",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		requireAWS()
		if awsAMIFlag == "" || awsNameFlag == "" {
			fmt.Fprintln(os.Stderr, "error: --ami and --name are required")
			os.Exit(1)
		}
		instType := awsTypeFlag
		if instType == "" {
			instType = "t2.micro"
		}

		// 1. Create key pair.
		keyName := "rift-" + awsNameFlag
		pemData, err := awsBackend.CreateKeyPair(keyName)
		if err != nil {
			internal.LogError(internal.ErrAWSCreateFailed, "", "creating key pair: %s", err)
			os.Exit(1)
		}
		pemPath := keyName + ".pem"
		if err := os.WriteFile(pemPath, []byte(pemData), 0600); err != nil {
			internal.LogError(internal.ErrAWSCreateFailed, "", "writing key file: %s", err)
			os.Exit(1)
		}
		fmt.Printf("key pair → %s\n", pemPath)

		// 2. Get default VPC + subnet.
		vpcID, err := awsBackend.GetDefaultVPC()
		if err != nil {
			internal.LogError(internal.ErrAWSCreateFailed, "", "finding default VPC: %s", err)
			os.Exit(1)
		}
		subnetID, err := awsBackend.GetFirstSubnet(vpcID)
		if err != nil {
			internal.LogError(internal.ErrAWSCreateFailed, "", "finding subnet: %s", err)
			os.Exit(1)
		}

		// 3. Security group.
		sgID, err := awsBackend.EnsureSecurityGroup(vpcID)
		if err != nil {
			internal.LogError(internal.ErrAWSCreateFailed, "", "security group: %s", err)
			os.Exit(1)
		}

		// 4. Launch instance.
		fmt.Printf("launching %s (%s)...\n", awsNameFlag, instType)
		instanceID, err := awsBackend.CreateInstance(awsAMIFlag, awsNameFlag, instType, keyName, sgID, subnetID)
		if err != nil {
			internal.LogError(internal.ErrAWSCreateFailed, "", "%s", err)
			os.Exit(1)
		}

		// 5. Wait for running.
		if err := awsBackend.WaitUntilRunning(instanceID, 5*time.Minute); err != nil {
			internal.LogError(internal.ErrAWSCreateFailed, instanceID, "waiting for running: %s", err)
			os.Exit(1)
		}

		// 6. Elastic IP.
		publicIP, err := awsBackend.AllocateAndAssociateEIP(instanceID)
		if err != nil {
			internal.LogError(internal.ErrAWSCreateFailed, instanceID, "elastic IP: %s", err)
			// Non-fatal — instance is running, just no EIP.
		}

		inst, _ := awsBackend.GetInstance(instanceID)
		if publicIP == "" {
			publicIP = inst.PublicIP
		}

		fmt.Println(strings.Repeat("-", 50))
		fmt.Printf("instance:  %s\n", instanceID)
		fmt.Printf("public IP: %s\n", publicIP)
		fmt.Printf("key file:  %s\n", pemPath)
		user := internal.GuessSSHUser(inst.Platform)
		if publicIP != "" {
			fmt.Printf("ssh:       ssh -i %s %s@%s\n", pemPath, user, publicIP)
		}
	},
}

// ---------------------------------------------------------------------------
// rift aws terminate
// ---------------------------------------------------------------------------

var awsTerminateCmd = &cobra.Command{
	Use:   "terminate <instance-id> [instance-id...]",
	Short: "Terminate EC2 instances",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		requireAWS()
		if !awsYesFlag {
			fmt.Fprintf(os.Stderr, "error: --yes is required to confirm termination\n")
			os.Exit(1)
		}
		// Release EIPs first.
		for _, id := range args {
			if err := awsBackend.ReleaseInstanceEIPs(id); err != nil {
				fmt.Fprintf(os.Stderr, "%s: warning: releasing EIP: %s\n", id, err)
			}
		}
		if err := awsBackend.TerminateInstances(args); err != nil {
			internal.LogError(internal.ErrAWSTermFailed, "", "%s", err)
			os.Exit(1)
		}
		for _, id := range args {
			fmt.Printf("%s → terminated\n", id)
		}
	},
}

// ---------------------------------------------------------------------------
// rift aws ssh
// ---------------------------------------------------------------------------

var awsSSHCmd = &cobra.Command{
	Use:   "ssh <instance-id>",
	Short: "SSH into an EC2 instance",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		requireAWS()
		inst, err := awsBackend.GetInstance(args[0])
		if err != nil {
			internal.LogError(internal.ErrAWSNotFound, args[0], "%s", err)
			os.Exit(1)
		}
		if inst.PublicIP == "" {
			fmt.Fprintln(os.Stderr, "error: instance has no public IP")
			os.Exit(1)
		}
		user := awsUserFlag
		if user == "" {
			user = internal.GuessSSHUser(inst.Platform)
		}
		pemPath := "rift-" + inst.KeyName + ".pem"
		// If that doesn't exist, try just keyname.pem.
		if _, err := os.Stat(pemPath); err != nil {
			pemPath = inst.KeyName + ".pem"
		}
		sshCmd := fmt.Sprintf("ssh -i %s %s@%s", pemPath, user, inst.PublicIP)
		fmt.Println(sshCmd)

		// Connect directly.
		c := exec.Command("ssh", "-i", pemPath, fmt.Sprintf("%s@%s", user, inst.PublicIP))
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			os.Exit(1)
		}
	},
}

// ---------------------------------------------------------------------------
// rift aws ip
// ---------------------------------------------------------------------------

var awsIPCmd = &cobra.Command{
	Use:   "ip <instance-id>",
	Short: "Show public and private IP of an EC2 instance",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		requireAWS()
		inst, err := awsBackend.GetInstance(args[0])
		if err != nil {
			internal.LogError(internal.ErrAWSNotFound, args[0], "%s", err)
			os.Exit(1)
		}
		pub := inst.PublicIP
		if pub == "" {
			pub = "(none)"
		}
		fmt.Printf("public:  %s\n", pub)
		fmt.Printf("private: %s\n", inst.PrivateIP)
	},
}

// ---------------------------------------------------------------------------
// Registration
// ---------------------------------------------------------------------------

func init() {
	rootCmd.AddCommand(awsCmd)
	awsCmd.AddCommand(awsListCmd)
	awsCmd.AddCommand(awsStartCmd)
	awsCmd.AddCommand(awsStopCmd)
	awsCmd.AddCommand(awsCreateCmd)
	awsCmd.AddCommand(awsTerminateCmd)
	awsCmd.AddCommand(awsSSHCmd)
	awsCmd.AddCommand(awsIPCmd)

	awsListCmd.Flags().BoolVar(&awsAllFlag, "all", false, "Include terminated instances")
	awsStopCmd.Flags().BoolVarP(&awsHardFlag, "hard", "H", false, "Force stop")
	awsTerminateCmd.Flags().BoolVarP(&awsYesFlag, "yes", "y", false, "Confirm termination")
	awsCreateCmd.Flags().StringVar(&awsAMIFlag, "ami", "", "AMI ID (required)")
	awsCreateCmd.Flags().StringVar(&awsNameFlag, "name", "", "Instance name (required)")
	awsCreateCmd.Flags().StringVar(&awsTypeFlag, "type", "t2.micro", "Instance type")
	awsCmd.PersistentFlags().StringVar(&awsRegionFlag, "region", "", "AWS region (overrides AWS_REGION from .env)")
	awsSSHCmd.Flags().StringVarP(&awsUserFlag, "user", "u", "", "SSH username (default: auto-detect from AMI)")
}
