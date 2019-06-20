package main

const (
	// This profile allows `unshare` syscall, to see the difference between this
	// profile and the default one run following command:
	// diff -u seccomp.go seccomp-weak.go
	weakSeccompConfig = `{
		"defaultAction": "SCMP_ACT_ERRNO",
		"archMap": [
			{
				"architecture": "SCMP_ARCH_X86_64",
				"subArchitectures": [
					"SCMP_ARCH_X86",
					"SCMP_ARCH_X32"
				]
			},
			{
				"architecture": "SCMP_ARCH_AARCH64",
				"subArchitectures": [
					"SCMP_ARCH_ARM"
				]
			},
			{
				"architecture": "SCMP_ARCH_MIPS64",
				"subArchitectures": [
					"SCMP_ARCH_MIPS",
					"SCMP_ARCH_MIPS64N32"
				]
			},
			{
				"architecture": "SCMP_ARCH_MIPS64N32",
				"subArchitectures": [
					"SCMP_ARCH_MIPS",
					"SCMP_ARCH_MIPS64"
				]
			},
			{
				"architecture": "SCMP_ARCH_MIPSEL64",
				"subArchitectures": [
					"SCMP_ARCH_MIPSEL",
					"SCMP_ARCH_MIPSEL64N32"
				]
			},
			{
				"architecture": "SCMP_ARCH_MIPSEL64N32",
				"subArchitectures": [
					"SCMP_ARCH_MIPSEL",
					"SCMP_ARCH_MIPSEL64"
				]
			},
			{
				"architecture": "SCMP_ARCH_S390X",
				"subArchitectures": [
					"SCMP_ARCH_S390"
				]
			}
		],
		"syscalls": [
			{
				"names": [
					"accept",
					"accept4",
					"access",
					"adjtimex",
					"alarm",
					"bind",
					"brk",
					"capget",
					"capset",
					"chdir",
					"chmod",
					"chown",
					"chown32",
					"clock_getres",
					"clock_gettime",
					"clock_nanosleep",
					"close",
					"connect",
					"copy_file_range",
					"creat",
					"dup",
					"dup2",
					"dup3",
					"epoll_create",
					"epoll_create1",
					"epoll_ctl",
					"epoll_ctl_old",
					"epoll_pwait",
					"epoll_wait",
					"epoll_wait_old",
					"eventfd",
					"eventfd2",
					"execve",
					"execveat",
					"exit",
					"exit_group",
					"faccessat",
					"fadvise64",
					"fadvise64_64",
					"fallocate",
					"fanotify_mark",
					"fchdir",
					"fchmod",
					"fchmodat",
					"fchown",
					"fchown32",
					"fchownat",
					"fcntl",
					"fcntl64",
					"fdatasync",
					"fgetxattr",
					"flistxattr",
					"flock",
					"fork",
					"fremovexattr",
					"fsetxattr",
					"fstat",
					"fstat64",
					"fstatat64",
					"fstatfs",
					"fstatfs64",
					"fsync",
					"ftruncate",
					"ftruncate64",
					"futex",
					"futimesat",
					"getcpu",
					"getcwd",
					"getdents",
					"getdents64",
					"getegid",
					"getegid32",
					"geteuid",
					"geteuid32",
					"getgid",
					"getgid32",
					"getgroups",
					"getgroups32",
					"getitimer",
					"getpeername",
					"getpgid",
					"getpgrp",
					"getpid",
					"getppid",
					"getpriority",
					"getrandom",
					"getresgid",
					"getresgid32",
					"getresuid",
					"getresuid32",
					"getrlimit",
					"get_robust_list",
					"getrusage",
					"getsid",
					"getsockname",
					"getsockopt",
					"get_thread_area",
					"gettid",
					"gettimeofday",
					"getuid",
					"getuid32",
					"getxattr",
					"inotify_add_watch",
					"inotify_init",
					"inotify_init1",
					"inotify_rm_watch",
					"io_cancel",
					"ioctl",
					"io_destroy",
					"io_getevents",
					"io_pgetevents",
					"ioprio_get",
					"ioprio_set",
					"io_setup",
					"io_submit",
					"ipc",
					"kill",
					"lchown",
					"lchown32",
					"lgetxattr",
					"link",
					"linkat",
					"listen",
					"listxattr",
					"llistxattr",
					"_llseek",
					"lremovexattr",
					"lseek",
					"lsetxattr",
					"lstat",
					"lstat64",
					"madvise",
					"memfd_create",
					"mincore",
					"mkdir",
					"mkdirat",
					"mknod",
					"mknodat",
					"mlock",
					"mlock2",
					"mlockall",
					"mmap",
					"mmap2",
					"mprotect",
					"mq_getsetattr",
					"mq_notify",
					"mq_open",
					"mq_timedreceive",
					"mq_timedsend",
					"mq_unlink",
					"mremap",
					"msgctl",
					"msgget",
					"msgrcv",
					"msgsnd",
					"msync",
					"munlock",
					"munlockall",
					"munmap",
					"nanosleep",
					"newfstatat",
					"_newselect",
					"open",
					"openat",
					"pause",
					"pipe",
					"pipe2",
					"poll",
					"ppoll",
					"prctl",
					"pread64",
					"preadv",
					"preadv2",
					"prlimit64",
					"pselect6",
					"pwrite64",
					"pwritev",
					"pwritev2",
					"read",
					"readahead",
					"readlink",
					"readlinkat",
					"readv",
					"recv",
					"recvfrom",
					"recvmmsg",
					"recvmsg",
					"remap_file_pages",
					"removexattr",
					"rename",
					"renameat",
					"renameat2",
					"restart_syscall",
					"rmdir",
					"rt_sigaction",
					"rt_sigpending",
					"rt_sigprocmask",
					"rt_sigqueueinfo",
					"rt_sigreturn",
					"rt_sigsuspend",
					"rt_sigtimedwait",
					"rt_tgsigqueueinfo",
					"sched_getaffinity",
					"sched_getattr",
					"sched_getparam",
					"sched_get_priority_max",
					"sched_get_priority_min",
					"sched_getscheduler",
					"sched_rr_get_interval",
					"sched_setaffinity",
					"sched_setattr",
					"sched_setparam",
					"sched_setscheduler",
					"sched_yield",
					"seccomp",
					"select",
					"semctl",
					"semget",
					"semop",
					"semtimedop",
					"send",
					"sendfile",
					"sendfile64",
					"sendmmsg",
					"sendmsg",
					"sendto",
					"setfsgid",
					"setfsgid32",
					"setfsuid",
					"setfsuid32",
					"setgid",
					"setgid32",
					"setgroups",
					"setgroups32",
					"setitimer",
					"setpgid",
					"setpriority",
					"setregid",
					"setregid32",
					"setresgid",
					"setresgid32",
					"setresuid",
					"setresuid32",
					"setreuid",
					"setreuid32",
					"setrlimit",
					"set_robust_list",
					"setsid",
					"setsockopt",
					"set_thread_area",
					"set_tid_address",
					"setuid",
					"setuid32",
					"setxattr",
					"shmat",
					"shmctl",
					"shmdt",
					"shmget",
					"shutdown",
					"sigaltstack",
					"signalfd",
					"signalfd4",
					"sigreturn",
					"socket",
					"socketcall",
					"socketpair",
					"splice",
					"stat",
					"stat64",
					"statfs",
					"statfs64",
					"statx",
					"symlink",
					"symlinkat",
					"sync",
					"sync_file_range",
					"syncfs",
					"sysinfo",
					"tee",
					"tgkill",
					"time",
					"timer_create",
					"timer_delete",
					"timerfd_create",
					"timerfd_gettime",
					"timerfd_settime",
					"timer_getoverrun",
					"timer_gettime",
					"timer_settime",
					"times",
					"tkill",
					"truncate",
					"truncate64",
					"ugetrlimit",
					"umask",
					"uname",
					"unlink",
					"unlinkat",
					"utime",
					"utimensat",
					"utimes",
					"vfork",
					"vmsplice",
					"wait4",
					"waitid",
					"waitpid",
					"write",
					"writev"
				],
				"action": "SCMP_ACT_ALLOW",
				"args": [],
				"comment": "",
				"includes": {},
				"excludes": {}
			},
			{
				"names": [
					"ptrace"
				],
				"action": "SCMP_ACT_ALLOW",
				"args": null,
				"comment": "",
				"includes": {
					"minKernel": "4.8"
				},
				"excludes": {}
			},
			{
				"names": [
					"personality"
				],
				"action": "SCMP_ACT_ALLOW",
				"args": [
					{
						"index": 0,
						"value": 0,
						"valueTwo": 0,
						"op": "SCMP_CMP_EQ"
					}
				],
				"comment": "",
				"includes": {},
				"excludes": {}
			},
			{
				"names": [
					"personality"
				],
				"action": "SCMP_ACT_ALLOW",
				"args": [
					{
						"index": 0,
						"value": 8,
						"valueTwo": 0,
						"op": "SCMP_CMP_EQ"
					}
				],
				"comment": "",
				"includes": {},
				"excludes": {}
			},
			{
				"names": [
					"personality"
				],
				"action": "SCMP_ACT_ALLOW",
				"args": [
					{
						"index": 0,
						"value": 131072,
						"valueTwo": 0,
						"op": "SCMP_CMP_EQ"
					}
				],
				"comment": "",
				"includes": {},
				"excludes": {}
			},
			{
				"names": [
					"personality"
				],
				"action": "SCMP_ACT_ALLOW",
				"args": [
					{
						"index": 0,
						"value": 131080,
						"valueTwo": 0,
						"op": "SCMP_CMP_EQ"
					}
				],
				"comment": "",
				"includes": {},
				"excludes": {}
			},
			{
				"names": [
					"personality"
				],
				"action": "SCMP_ACT_ALLOW",
				"args": [
					{
						"index": 0,
						"value": 4294967295,
						"valueTwo": 0,
						"op": "SCMP_CMP_EQ"
					}
				],
				"comment": "",
				"includes": {},
				"excludes": {}
			},
			{
				"names": [
					"sync_file_range2"
				],
				"action": "SCMP_ACT_ALLOW",
				"args": [],
				"comment": "",
				"includes": {
					"arches": [
						"ppc64le"
					]
				},
				"excludes": {}
			},
			{
				"names": [
					"arm_fadvise64_64",
					"arm_sync_file_range",
					"sync_file_range2",
					"breakpoint",
					"cacheflush",
					"set_tls"
				],
				"action": "SCMP_ACT_ALLOW",
				"args": [],
				"comment": "",
				"includes": {
					"arches": [
						"arm",
						"arm64"
					]
				},
				"excludes": {}
			},
			{
				"names": [
					"arch_prctl"
				],
				"action": "SCMP_ACT_ALLOW",
				"args": [],
				"comment": "",
				"includes": {
					"arches": [
						"amd64",
						"x32"
					]
				},
				"excludes": {}
			},
			{
				"names": [
					"modify_ldt"
				],
				"action": "SCMP_ACT_ALLOW",
				"args": [],
				"comment": "",
				"includes": {
					"arches": [
						"amd64",
						"x32",
						"x86"
					]
				},
				"excludes": {}
			},
			{
				"names": [
					"s390_pci_mmio_read",
					"s390_pci_mmio_write",
					"s390_runtime_instr"
				],
				"action": "SCMP_ACT_ALLOW",
				"args": [],
				"comment": "",
				"includes": {
					"arches": [
						"s390",
						"s390x"
					]
				},
				"excludes": {}
			},
			{
				"names": [
					"open_by_handle_at"
				],
				"action": "SCMP_ACT_ALLOW",
				"args": [],
				"comment": "",
				"includes": {
					"caps": [
						"CAP_DAC_READ_SEARCH"
					]
				},
				"excludes": {}
			},
			{
				"names": [
					"bpf",
					"clone",
					"fanotify_init",
					"lookup_dcookie",
					"mount",
					"name_to_handle_at",
					"perf_event_open",
					"quotactl",
					"setdomainname",
					"sethostname",
					"setns",
					"syslog",
					"umount",
					"umount2"
				],
				"action": "SCMP_ACT_ALLOW",
				"args": [],
				"comment": "",
				"includes": {
					"caps": [
						"CAP_SYS_ADMIN"
					]
				},
				"excludes": {}
			},
			{
				"names": [
					"unshare"
				],
				"action": "SCMP_ACT_ALLOW",
				"args": [],
				"comment": "",
				"includes": {},
				"excludes": {}
			},
			{
				"names": [
					"clone"
				],
				"action": "SCMP_ACT_ALLOW",
				"args": [
					{
						"index": 0,
						"value": 2114060288,
						"valueTwo": 0,
						"op": "SCMP_CMP_MASKED_EQ"
					}
				],
				"comment": "",
				"includes": {},
				"excludes": {
					"caps": [
						"CAP_SYS_ADMIN"
					],
					"arches": [
						"s390",
						"s390x"
					]
				}
			},
			{
				"names": [
					"clone"
				],
				"action": "SCMP_ACT_ALLOW",
				"args": [
					{
						"index": 1,
						"value": 2114060288,
						"valueTwo": 0,
						"op": "SCMP_CMP_MASKED_EQ"
					}
				],
				"comment": "s390 parameter ordering for clone is different",
				"includes": {
					"arches": [
						"s390",
						"s390x"
					]
				},
				"excludes": {
					"caps": [
						"CAP_SYS_ADMIN"
					]
				}
			},
			{
				"names": [
					"reboot"
				],
				"action": "SCMP_ACT_ALLOW",
				"args": [],
				"comment": "",
				"includes": {
					"caps": [
						"CAP_SYS_BOOT"
					]
				},
				"excludes": {}
			},
			{
				"names": [
					"chroot"
				],
				"action": "SCMP_ACT_ALLOW",
				"args": [],
				"comment": "",
				"includes": {
					"caps": [
						"CAP_SYS_CHROOT"
					]
				},
				"excludes": {}
			},
			{
				"names": [
					"delete_module",
					"init_module",
					"finit_module",
					"query_module"
				],
				"action": "SCMP_ACT_ALLOW",
				"args": [],
				"comment": "",
				"includes": {
					"caps": [
						"CAP_SYS_MODULE"
					]
				},
				"excludes": {}
			},
			{
				"names": [
					"acct"
				],
				"action": "SCMP_ACT_ALLOW",
				"args": [],
				"comment": "",
				"includes": {
					"caps": [
						"CAP_SYS_PACCT"
					]
				},
				"excludes": {}
			},
			{
				"names": [
					"kcmp",
					"process_vm_readv",
					"process_vm_writev",
					"ptrace"
				],
				"action": "SCMP_ACT_ALLOW",
				"args": [],
				"comment": "",
				"includes": {
					"caps": [
						"CAP_SYS_PTRACE"
					]
				},
				"excludes": {}
			},
			{
				"names": [
					"iopl",
					"ioperm"
				],
				"action": "SCMP_ACT_ALLOW",
				"args": [],
				"comment": "",
				"includes": {
					"caps": [
						"CAP_SYS_RAWIO"
					]
				},
				"excludes": {}
			},
			{
				"names": [
					"settimeofday",
					"stime",
					"clock_settime"
				],
				"action": "SCMP_ACT_ALLOW",
				"args": [],
				"comment": "",
				"includes": {
					"caps": [
						"CAP_SYS_TIME"
					]
				},
				"excludes": {}
			},
			{
				"names": [
					"vhangup"
				],
				"action": "SCMP_ACT_ALLOW",
				"args": [],
				"comment": "",
				"includes": {
					"caps": [
						"CAP_SYS_TTY_CONFIG"
					]
				},
				"excludes": {}
			},
			{
				"names": [
					"get_mempolicy",
					"mbind",
					"set_mempolicy"
				],
				"action": "SCMP_ACT_ALLOW",
				"args": [],
				"comment": "",
				"includes": {
					"caps": [
						"CAP_SYS_NICE"
					]
				},
				"excludes": {}
			},
			{
				"names": [
					"syslog"
				],
				"action": "SCMP_ACT_ALLOW",
				"args": [],
				"comment": "",
				"includes": {
					"caps": [
						"CAP_SYSLOG"
					]
				},
				"excludes": {}
			}
		]
	}`
)
