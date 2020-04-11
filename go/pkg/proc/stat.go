package proc

import (
	"errors"
	"fmt"
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
	Pid                 int16         // (1) The process id.
	Comm                string        // (2) The filename of the executable, in parentheses.
	State               StatTaskState // (3) process state
	PPid                int16         // (4) The PID of the parent of this process.
	PGrp                int16         // (5) The process group ID of the process.
	Session             int16         // (6) The session ID of the process.
	TTYNR               int16         // (7) The controlling terminal of the process.
	TPGid               int16         // (8) The ID of the foreground process group of the controlling terminal of the process.
	TaskFlags           int32         // (9) The kernel flags word of the process
	MinFlt              uint32        // (10) The number of minor faults the process has made which have not required loading a memory page from disk.
	CMinFlt             uint32        // (11) The number of minor faults that the process's waited-for children have made.
	MajFlt              uint32        // (12) The number of major faults the process has made which have required loading a memory page from disk.
	CMajFlt             uint32        // (13) The number of major faults that the process's waited-for children have made.
	UTime               uint32        // (14) user mode jiffies
	Stime               uint32        // (15) kernel mode jiffies
	CUTime              int32         // (16) user mode jiffies with childs
	CSTime              int32         // (17) kernel mode jiffies with childs
	Priority            int32         // (18) (Explanation for Linux 2.6) For processes running a real-time scheduling policy
	Nice                int32         // (19) The nice value (see setpriority(2)), a value in the range 19 (low priority) to -20 (high priority).
	NumThreads          int32         // (20) number of threads in this process (since Linux 2.6).
	ItRealValue         int32         // (21) The time in jiffies before the next SIGALRM is sent  to the process due to an interval timer.
	StartTime           uint64        // (22) The time the process started after system boot.
	VSize               uint32        // (23) Virtual memory size
	RSS                 int32         // (24) Resident Set Size: number of pages the process has in real memory.
	RSSLim              uint32        // (25) Current limit in bytes on the rss of the process
	StartCode           uint32        // (26) The address above which program text can run.
	EndCode             uint32        // (27) The address below which program text can run.
	StartStack          uint32        // (28) The address of the start (i.e., bottom) of the stack
	KStkEsp             uint32        // (29) The current value of ESP (stack pointer), as found in the kernel stack page for the process.
	KStkEip             uint32        // (30) The current EIP (instruction pointer).
	TaskPendingSig      uint32        // (31) The bitmap of pending signals, displayed as a decimal number.
	TaskBlockSig        uint32        // (32) The bitmap of blocked signals, displayed as a decimal number.
	SigIgnoreSig        uint32        // (33) The bitmap of ignored signals, displayed as a decimal number.
	SigCatchSig         uint32        // (34) The bitmap of caught signals, displayed as a decimal number.
	WChan               uint32        // (35) This is the "channel" in which the process is waiting.
	NSwap               uint32        // (36) Number of pages swapped (not maintained).
	CNSwap              uint32        // (37) Cumulative nswap for child processes (not maintained).
	ExitSignal          int16         // (38) Signal to be sent to parent when we die. (since Linux 2.1.22)
	Processor           int16         // (39) CPU number last executed on. (since Linux 2.2.8)
	RtPriority          uint16        // (40) Real-time scheduling priority
	Policy              uint16        // (41) Scheduling policy (see sched_setscheduler(2)).
	DelayacctBlkioTicks uint64        // (42) Aggregated block I/O delays, measured in clock ticks (centiseconds).
	GuestTime           uint32        // (43)  Guest time of the process
	CGuestTime          int32         // (44) Guest time of the process's children
	StartData           uint32        // (45) Address above which program initialized and uninitialized (BSS) data are placed.
	EndData             uint32        // (46) Address below which program initialized and uninitialized (BSS) data are placed.
	StartBrk            uint32        // (47) Address above which program heap can be expanded with brk(2).
	ArgStart            uint32        // (48) Address below program command-line arguments (argv) are placed.
	ArgEnd              uint32        // (49) Address above which program environment is placed.
	EnvStart            uint32        // (50) Address below which program environment is placed.
	EnvEnd              uint32        // (51) Address below which program environment is placed.
	ExitCode            int16         // (52) The thread's exit status in the form reported by waitpid(2).
}

var fieldParseMap = map[int]func(field string, sf *StatField) (err error){
	1: func(field string, sf *StatField) (err error) {
		i16, err := int16Cover(field)
		if err != nil {
			return err
		}

		sf.Pid = i16
		return nil
	},

	2: func(field string, sf *StatField) (err error) {
		sf.Comm = field
		return nil
	},

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

	4: func(field string, sf *StatField) (err error) {
		sf.PPid, err = int16Cover(field)
		return
	},

	5: func(field string, sf *StatField) (err error) {
		sf.PGrp, err = int16Cover(field)
		return
	},

	6: func(field string, sf *StatField) (err error) {
		sf.Session, err = int16Cover(field)
		return
	},

	7: func(field string, sf *StatField) (err error) {
		sf.TTYNR, err = int16Cover(field)
		return
	},

	8: func(field string, sf *StatField) (err error) {
		sf.TPGid, err = int16Cover(field)
		return
	},

	9: func(field string, sf *StatField) (err error) {
		sf.TaskFlags, err = int32Cover(field)
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
		i64, err := strconv.ParseInt(s, 10, 16)
		if err != nil {
			return 0, err
		}

		return uint16(i64), nil
	}

	int32Cover = func(s string) (int32, error) {
		i64, err := strconv.ParseInt(s, 10, 16)
		if err != nil {
			return 0, err
		}

		return int32(i64), nil
	}

	uint32Cover = func(s string) (uint32, error) {
		i64, err := strconv.ParseInt(s, 10, 16)
		if err != nil {
			return 0, err
		}

		return uint32(i64), nil
	}

	uint64Cover = func(s string) (uint64, error) {
		i64, err := strconv.ParseInt(s, 10, 16)
		if err != nil {
			return 0, err
		}

		return uint64(i64), nil
	}
)

func (sf *StatField) fill(field string, progress int) error {
	if parseFunc, ok := fieldParseMap[progress]; ok {
		if err := parseFunc(field, sf); err != nil {
			return err
		}
	}

	return errors.New("")
}

type StatStream struct {
	data *[]byte
	cur  int
	end  int
}

func (ss *StatStream) isEOF() bool {
	if ss.cur == ss.end {
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

func (ss *StatStream) next() string {
	buf := make([]byte, 0, 0)
	for {
		if ss.isEOF() {
			return ""
		}

		if ss.isSpace() {
			ss.cur++
			break
		}

		buf = append(buf, (*ss.data)[ss.cur])
		ss.cur++
	}

	return string(buf)
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

	pos := 1
	for {
		field := s.stream.next()
		fmt.Printf(" %s", field)
		if field == "" {
			break
		}
		s.fields.fill(field, pos)
		pos++
	}

	return s.fields, nil
}
