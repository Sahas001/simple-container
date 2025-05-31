package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"
)

func run() {
	fmt.Printf("Running %v as PID %d\n", os.Args[2:], os.Getpid())
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID,
	}

	if err := cmd.Run(); err != nil {
		fmt.Println("Error running the /proc/self/exe command", err)
		os.Exit(1)
	}
}

func child() {
	fmt.Printf("Running %v as PID %d\n", os.Args[2:], os.Getpid())

	// set hostname of the new UTS namespace
	if err := syscall.Sethostname([]byte("HMcontainer")); err != nil {
		fmt.Println("Error setting hostname:", err)
		os.Exit(1)
	}

	if err := syscall.Chroot("/home/sahas/Projects/Go/container/rootfs"); err != nil {
		fmt.Println("Error changing root directory:", err)
		os.Exit(1)
	}

	if err := syscall.Chdir("/"); err != nil {
		fmt.Println("Error changing directory:", err)
		os.Exit(1)
	}

	// mount the proc filesystem for managing processes

	if err := syscall.Mount("proc", "proc", "proc", 0, ""); err != nil {
		fmt.Println("Error mounting proc filesystem:", err)
		os.Exit(1)
	}

	userName := "guest"
	u, err := user.Lookup(userName)
	if err != nil {
		fmt.Println("Error looking up user:", err)
		os.Exit(1)
	}

	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		fmt.Println("Error converting UID:", err)
		os.Exit(1)
	}

	gid, err := strconv.Atoi(u.Gid)
	if err != nil {
		fmt.Println("Error converting GID:", err)
		os.Exit(1)
	}

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "PATH=/bin:/usr/bin:/sbin:/usr/sbin")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid:         uint32(uid),
			Gid:         uint32(gid),
			NoSetGroups: true,
		},
	}

	if err := cmd.Run(); err != nil {
		fmt.Println("Error running the child command in the new namespace:", err)
		os.Exit(1)
	}

}

func main() {
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		fmt.Println("Usage: ./container run <command> [args]")
		os.Exit(1)
	}
}
