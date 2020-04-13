package proc

import (
	"errors"
	"io/ioutil"
	"strconv"
)

type StatTaskState int8

const (
	Running       StatTaskState = 'R' // Running
	Sleeping      StatTaskState = 'S' // Sleeping in an interruptible wait
	DiskSleep     StatTaskState = 'D' //  Waiting in uninterruptible disk sleep
	Zombie        StatTaskState = 'Z' //  Waiting in uninterruptible disk sleep
	Stopped       StatTaskState = 'T' // Stopped (on a signal) or (before Linux 2.6.33) trace stopped
	TracingStop   StatTaskState = 't' // Tracing stop (Linux 2.6.33 onward)
	Dead          StatTaskState = 'X' // Dead (from Linux 2.6.0 onward)
	Dead2633To313 StatTaskState = 'x' // Dead (Linux 2.6.33 to 3.13 only)
	Wakekill      StatTaskState = 'K' // Wakekill (Linux 2.6.33 to 3.13 only)
	PagingWaking  StatTaskState = 'W' // Waking (Linux 2.6.33 to 3.13 only)
	Parked        StatTaskState = 'P' // Parked (Linux 3.9 to 3.13 only)
)

// Stat by https://github.com/torvalds/linux/blob/c0cc271173b2e1c2d8d0ceaef14e4dfa79eefc0d/fs/proc/array.c#L430
// define by http://man7.org/linux/man-pages/man5/proc.5.html
type StatField struct {
	Pid                 int32         // (1) The process id.
	Comm                string        // (2) The filename of the executable, in parentheses.
	State               StatTaskState // (3) process state
	PPid                int32         // (4) The PID of the parent of this process.
	PGrp                int32         // (5) The process group ID of the process.
	Session             int32         // (6) The session ID of the process.
	TTYNR               int32         // (7) The controlling terminal of the process.
	TPGid               int32         // (8) The ID of the foreground process group of the controlling terminal of the process.
	TaskFlags           uint32        // (9) The kernel flags word of the process
	MinFlt              uint64        // (10) The number of minor faults the process has made which have not required loading a memory page from disk.
	CMinFlt             uint64        // (11) The number of minor faults that the process's waited-for children have made.
	MajFlt              uint64        // (12) The number of major faults the process has made which have required loading a memory page from disk.
	CMajFlt             uint64        // (13) The number of major faults that the process's waited-for children have made.
	UTime               uint64        // (14) user mode jiffies
	STime               uint64        // (15) kernel mode jiffies
	CUTime              int64         // (16) user mode jiffies with childs
	CSTime              int64         // (17) kernel mode jiffies with childs
	Priority            int64         // (18) (Explanation for Linux 2.6) For processes running a real-time scheduling policy
	Nice                int64         // (19) The nice value (see setpriority(2)), a value in the range 19 (low priority) to -20 (high priority).
	NumThreads          int64         // (20) number of threads in this process (since Linux 2.6).
	ItRealValue         int64         // (21) The time in jiffies before the next SIGALRM is sent  to the process due to an interval timer.
	StartTime           uint64        // (22) The time the process started after system boot.
	VSize               uint64        // (23) Virtual memory size
	RSS                 uint64        // (24) Resident Set Size: number of pages the process has in real memory.
	RSSLim              uint64        // (25) Current limit in bytes on the rss of the process
	StartCode           uint64        // (26) The address above which program text can run.
	EndCode             uint64        // (27) The address below which program text can run.
	StartStack          uint64        // (28) The address of the start (i.e., bottom) of the stack
	KStkEsp             uint64        // (29) The current value of ESP (stack pointer), as found in the kernel stack page for the process.
	KStkEip             uint64        // (30) The current EIP (instruction pointer).
	TaskPendingSig      uint64        // (31) The bitmap of pending signals, displayed as a decimal number.
	TaskBlockSig        uint64        // (32) The bitmap of blocked signals, displayed as a decimal number.
	SigIgnoreSig        uint64        // (33) The bitmap of ignored signals, displayed as a decimal number.
	SigCatchSig         uint64        // (34) The bitmap of caught signals, displayed as a decimal number.
	WChan               uint64        // (35) This is the "channel" in which the process is waiting.
	NSwap               uint64        // (36) Number of pages swapped (not maintained).
	CNSwap              uint64        // (37) Cumulative nswap for child processes (not maintained).
	ExitSignal          int32         // (38) Signal to be sent to parent when we die. (since Linux 2.1.22)
	Processor           int32         // (39) CPU number last executed on. (since Linux 2.2.8)
	RtPriority          uint32        // (40) Real-time scheduling priority
	Policy              uint32        // (41) Scheduling policy (see sched_setscheduler(2)).
	DelayacctBlkioTicks uint64        // (42) Aggregated block I/O delays, measured in clock ticks (centiseconds).
	GuestTime           uint64        // (43)  Guest time of the process
	CGuestTime          int64         // (44) Guest time of the process's children
	StartData           uint64        // (45) Address above which program initialized and uninitialized (BSS) data are placed.
	EndData             uint64        // (46) Address below which program initialized and uninitialized (BSS) data are placed.
	StartBrk            uint64        // (47) Address above which program heap can be expanded with brk(2).
	ArgStart            uint64        // (48) Address below program command-line arguments (argv) are placed.
	ArgEnd              uint64        // (49) Address above which program environment is placed.
	EnvStart            uint64        // (50) Address below which program environment is placed.
	EnvEnd              uint64        // (51) Address below which program environment is placed.
	ExitCode            int32         // (52) The thread's exit status in the form reported by waitpid(2).
}

var fieldParseMap = map[int]func(field string, sf *StatField) (err error){
	// pid
	1: func(field string, sf *StatField) (err error) {
		sf.Pid, err = int32Cover(field)
		return
	},

	// comm
	2: func(field string, sf *StatField) (err error) {
		sf.Comm = field
		return nil
	},

	// state
	3: func(field string, sf *StatField) (err error) {
		data := []byte(field)
		switch StatTaskState(data[0]) {
		case Running:
			sf.State = Running
		case Sleeping:
			sf.State = Sleeping
		case DiskSleep:
			sf.State = DiskSleep
		case Zombie:
			sf.State = Zombie
		case Stopped:
			sf.State = Stopped
		case TracingStop:
			sf.State = TracingStop
		case PagingWaking:
			sf.State = PagingWaking
		case Dead:
			sf.State = Dead
		case Dead2633To313:
			sf.State = Dead2633To313
		case Wakekill:
			sf.State = Wakekill
		case Parked:
			sf.State = Parked
		default:
			sf.State = StatTaskState(data[0])
		}

		return nil
	},

	// ppid
	4: func(field string, sf *StatField) (err error) {
		sf.PPid, err = int32Cover(field)
		return
	},

	// pgrp
	5: func(field string, sf *StatField) (err error) {
		sf.PGrp, err = int32Cover(field)
		return
	},

	// session
	6: func(field string, sf *StatField) (err error) {
		sf.Session, err = int32Cover(field)
		return
	},

	// ttynr
	7: func(field string, sf *StatField) (err error) {
		sf.TTYNR, err = int32Cover(field)
		return
	},

	// tpgid
	8: func(field string, sf *StatField) (err error) {
		sf.TPGid, err = int32Cover(field)
		return
	},

	// taskflags
	9: func(field string, sf *StatField) (err error) {
		sf.TaskFlags, err = uint32Cover(field)
		return
	},

	// minflt
	10: func(field string, sf *StatField) (err error) {
		sf.MinFlt, err = uint64Cover(field)
		return
	},

	// cminflt
	11: func(field string, sf *StatField) (err error) {
		sf.CMinFlt, err = uint64Cover(field)
		return
	},

	// majflt
	12: func(field string, sf *StatField) (err error) {
		sf.MajFlt, err = uint64Cover(field)
		return
	},

	// cmajflt
	13: func(field string, sf *StatField) (err error) {
		sf.MajFlt, err = uint64Cover(field)
		return
	},

	// utime
	14: func(field string, sf *StatField) (err error) {
		sf.UTime, err = uint64Cover(field)
		return
	},

	// stime
	15: func(field string, sf *StatField) (err error) {
		sf.STime, err = uint64Cover(field)
		return
	},

	// cutime
	16: func(field string, sf *StatField) (err error) {
		sf.CUTime, err = int64Cover(field)
		return
	},

	// cstime
	17: func(field string, sf *StatField) (err error) {
		sf.CSTime, err = int64Cover(field)
		return
	},

	// priority
	18: func(field string, sf *StatField) (err error) {
		sf.Priority, err = int64Cover(field)
		return
	},

	// nice
	19: func(field string, sf *StatField) (err error) {
		sf.Nice, err = int64Cover(field)
		return
	},

	// numthreads
	20: func(field string, sf *StatField) (err error) {
		sf.NumThreads, err = int64Cover(field)
		return
	},

	// itrealvalue
	21: func(field string, sf *StatField) (err error) {
		sf.ItRealValue, err = int64Cover(field)
		return
	},

	// starttime
	22: func(field string, sf *StatField) (err error) {
		sf.StartTime, err = uint64Cover(field)
		return
	},

	// vszie
	23: func(field string, sf *StatField) (err error) {
		sf.VSize, err = uint64Cover(field)
		return
	},

	// rss
	24: func(field string, sf *StatField) (err error) {
		sf.RSS, err = uint64Cover(field)
		return
	},

	// rsslim
	25: func(field string, sf *StatField) (err error) {
		sf.RSSLim, err = uint64Cover(field)
		return
	},

	// startcode
	26: func(field string, sf *StatField) (err error) {
		sf.StartCode, err = uint64Cover(field)
		return
	},

	// endcode
	27: func(field string, sf *StatField) (err error) {
		sf.EndCode, err = uint64Cover(field)
		return
	},

	// startstack
	28: func(field string, sf *StatField) (err error) {
		sf.StartStack, err = uint64Cover(field)
		return
	},

	// kstkesp
	29: func(field string, sf *StatField) (err error) {
		sf.KStkEsp, err = uint64Cover(field)
		return
	},

	// kstkeip
	30: func(field string, sf *StatField) (err error) {
		sf.KStkEip, err = uint64Cover(field)
		return
	},

	// taskpendingsig
	31: func(field string, sf *StatField) (err error) {
		sf.TaskPendingSig, err = uint64Cover(field)
		return
	},

	// taskblocksig
	32: func(field string, sf *StatField) (err error) {
		sf.TaskBlockSig, err = uint64Cover(field)
		return
	},

	// sigignoresig
	33: func(field string, sf *StatField) (err error) {
		sf.SigIgnoreSig, err = uint64Cover(field)
		return
	},

	// sigcatchsig
	34: func(field string, sf *StatField) (err error) {
		sf.SigCatchSig, err = uint64Cover(field)
		return
	},

	// wchan
	35: func(field string, sf *StatField) (err error) {
		sf.WChan, err = uint64Cover(field)
		return
	},

	// nswap
	36: func(field string, sf *StatField) (err error) {
		sf.NSwap, err = uint64Cover(field)
		return
	},

	// cnswap
	37: func(field string, sf *StatField) (err error) {
		sf.CNSwap, err = uint64Cover(field)
		return
	},

	// exitsignal
	38: func(field string, sf *StatField) (err error) {
		sf.ExitSignal, err = int32Cover(field)
		return
	},

	// processor
	39: func(field string, sf *StatField) (err error) {
		sf.Processor, err = int32Cover(field)
		return
	},

	// rtpriority
	40: func(field string, sf *StatField) (err error) {
		sf.RtPriority, err = uint32Cover(field)
		return
	},

	// policy
	41: func(field string, sf *StatField) (err error) {
		sf.Policy, err = uint32Cover(field)
		return
	},

	// delayacctblkioticks
	42: func(field string, sf *StatField) (err error) {
		sf.DelayacctBlkioTicks, err = uint64Cover(field)
		return
	},

	// guesttime
	43: func(field string, sf *StatField) (err error) {
		sf.GuestTime, err = uint64Cover(field)
		return
	},

	// cguesttime
	44: func(field string, sf *StatField) (err error) {
		sf.CGuestTime, err = int64Cover(field)
		return
	},

	// startdata
	45: func(field string, sf *StatField) (err error) {
		sf.StartData, err = uint64Cover(field)
		return
	},

	// enddata
	46: func(field string, sf *StatField) (err error) {
		sf.EndData, err = uint64Cover(field)
		return
	},

	// startbrk
	47: func(field string, sf *StatField) (err error) {
		sf.StartBrk, err = uint64Cover(field)
		return
	},

	// argstart
	48: func(field string, sf *StatField) (err error) {
		sf.ArgStart, err = uint64Cover(field)
		return
	},

	// argend
	49: func(field string, sf *StatField) (err error) {
		sf.ArgEnd, err = uint64Cover(field)
		return
	},

	// envstart
	50: func(field string, sf *StatField) (err error) {
		sf.EnvStart, err = uint64Cover(field)
		return
	},

	// envend
	51: func(field string, sf *StatField) (err error) {
		sf.EnvEnd, err = uint64Cover(field)
		return
	},

	// exitcode
	52: func(field string, sf *StatField) (err error) {
		sf.ExitCode, err = int32Cover(field)
		return
	},
}

var (
	int16Cover = func(s string) (int16, error) {
		i64, err := strconv.ParseInt(s, 10, 16)
		if err != nil {
			return 0, err
		}

		return int16(i64), nil
	}

	uint16Cover = func(s string) (uint16, error) {
		ui64, err := strconv.ParseUint(s, 10, 16)
		if err != nil {
			return 0, err
		}

		return uint16(ui64), nil
	}

	int32Cover = func(s string) (int32, error) {
		i64, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return 0, err
		}

		return int32(i64), nil
	}

	uint32Cover = func(s string) (uint32, error) {
		ui64, err := strconv.ParseUint(s, 10, 32)
		if err != nil {
			return 0, err
		}

		return uint32(ui64), nil
	}

	int64Cover = func(s string) (int64, error) {
		return strconv.ParseInt(s, 10, 64)
	}

	uint64Cover = func(s string) (uint64, error) {
		return strconv.ParseUint(s, 10, 64)
	}
)

func (sf *StatField) fill(field string, progress int) error {
	if parseFunc, ok := fieldParseMap[progress]; ok {
		if err := parseFunc(field, sf); err != nil {
			return err
		}
	}

	return nil
}

var erreof = errors.New("stat stream read end of file")

type StatStream struct {
	data *[]byte
	cur  int
	end  int
}

func (ss *StatStream) isEOF() bool {
	if (*ss.data)[ss.cur] == '\n' {
		return true
	}
	return false
}

func (ss *StatStream) isSpace() bool {
	if (*ss.data)[ss.cur] == ' ' {
		return true
	}
	return false
}

func (ss *StatStream) next() (string, error) {
	buf := make([]byte, 0, 0)
	for {
		if ss.isEOF() {
			if len(buf) == 0 {
				return "", erreof
			}

			return string(buf), erreof
		}

		if ss.isSpace() {
			ss.cur++
			break
		}

		buf = append(buf, (*ss.data)[ss.cur])
		ss.cur++
	}
	return string(buf), nil
}

type Stat struct {
	pf     string
	stream *StatStream
	fields *StatField
}

func NewStat(proc string) *Stat {
	stat := &Stat{
		pf:     proc,
		stream: new(StatStream),
		fields: new(StatField),
	}

	return stat
}

func (s *Stat) Parse() (*StatField, error) {
	data, err := ioutil.ReadFile(s.pf)
	if err != nil {
		return nil, err
	}

	s.stream.data = &data
	s.stream.cur = 0
	s.stream.end = len(*s.stream.data)
	parseProgressIndex := 1
	for {
		field, _ := s.stream.next()
		if field == "" {
			break
		}
		if err = s.fields.fill(field, parseProgressIndex); err != nil {
			return nil, err
		}

		parseProgressIndex++
	}

	return s.fields, nil
}
