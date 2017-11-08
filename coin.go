package main

/*
This software was developed by employees of the National Institute of
Standards and Technology (NIST), an agency of the Federal Government.
Pursuant to title 17 United States Code Section 105, works of NIST
employees are not subject to copyright protection in the United States
and are considered to be in the public domain. Permission to freely
use, copy, modify, and distribute this software and its documentation
without fee is hereby granted, provided that this notice and disclaimer
of warranty appears in all copies.

THE SOFTWARE IS PROVIDED 'AS IS' WITHOUT ANY WARRANTY OF ANY KIND,
EITHER EXPRESSED, IMPLIED, OR STATUTORY, INCLUDING, BUT NOT LIMITED TO,
ANY WARRANTY THAT THE SOFTWARE WILL CONFORM TO SPECIFICATIONS, ANY
IMPLIED WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE,
AND FREEDOM FROM INFRINGEMENT, AND ANY WARRANTY THAT THE DOCUMENTATION
WILL CONFORM TO THE SOFTWARE, OR ANY WARRANTY THAT THE SOFTWARE WILL BE
ERROR FREE. IN NO EVENT SHALL NIST BE LIABLE FOR ANY DAMAGES, INCLUDING,
BUT NOT LIMITED TO, DIRECT, INDIRECT, SPECIAL OR CONSEQUENTIAL DAMAGES,
ARISING OUT OF, RESULTING FROM, OR IN ANY WAY CONNECTED WITH THIS SOFTWARE,
WHETHER OR NOT BASED UPON WARRANTY, CONTRACT, TORT, OR OTHERWISE, WHETHER
OR NOT INJURY WAS SUSTAINED BY PERSONS OR PROPERTY OR OTHERWISE, AND
WHETHER OR NOT LOSS WAS SUSTAINED FROM, OR AROSE OUT OF THE RESULTS OF,
OR USE OF, THE SOFTWARE OR SERVICES PROVIDED HEREUNDER.
*/

/*
  Stephen Nightingale
  night@nist.gov
  NIST, Information Technology Laboratory
  March 14, 2017
*/


import (
  "bufio"
  "encoding/base64"
  "encoding/json"
  "flag"
  "fmt"
  "math/rand"
  "net"
  "os"
  "currency/himitsu"
  "currency/methods"
  "currency/newtxes"
  "currency/structures"
  "currency/pairs"
  "strconv"
  "strings"
  "time"
)


// This is the client for the Scrooge masterpay system.
// Commands include:
// - Balance (get my balance)
// - PayCoins (Pay from my balance to Payee)
// - Quit (Remotely close the masterpay listener. Only Scrooge can do this).
// Usage:
// coin Balance conf/<user.conf>
// coin PayCoins conf/<user.conf> <CoinCount> keys/<payee.pub>
// coin Quit conf/scrooge.conf

// This is the package scope for all balances received.
var ballons = map[string]int{}
var baltime = map[string]int64{}
var allaccounts = map[string][]structures.Coin{}
var rollover = 60000

// Package scope for pre-populating a random ramp.
var randramp = []int{}
var nexrand = 0

var row, col = 0, 0
// var pairs = [][]string{
// {"alice/bob", "bob/chaz", "chaz/dave", "dave/ellie", "ellie/fiona", "fiona/gary", "gary/helen", "helen/alice"},
// {"alice/chaz", "bob/dave", "chaz/ellie", "dave/fiona", "ellie/gary", "fiona/helen", "gary/alice", "helen/bob"},
// {"alice/dave", "bob/ellie", "chaz/fiona", "dave/gary", "ellie/helen", "fiona/alice", "gary/bob", "helen/chaz"},
// {"alice/ellie", "bob/fiona", "chaz/gary", "dave/helen", "ellie/alice", "fiona/bob", "gary/chaz", "helen/dave"},
// {"alice/fiona", "bob/gary", "chaz/helen", "dave/alice", "ellie/bob", "fiona/chaz", "gary/dave", "helen/ellie"},
// {"alice/gary", "bob/helen", "chaz/alice", "dave/bob", "ellie/chaz", "fiona/dave", "gary/ellie", "helen/fiona"},
// {"alice/helen", "bob/alice", "chaz/bob", "dave/chaz", "ellie/dave", "fiona/ellie", "gary/fiona", "helen/gary"},
//  } // endpairs.
//

// Package scope for the Message stack.
type Message struct {
  Mtype string
  Mmap map[string]string
  Mcoins []structures.Coin
} // type Message.

type Stack []Message
var stack = Stack{}

func (s Stack) Empty() bool { return len(s) == 0 }
func (s Stack) Peek() Message { return s[len(s) - 1] }
func (s *Stack) Put(m Message) { (*s) = append((*s), m) }
func (s *Stack) Pop () Message {
  d := (*s)[len(*s) - 1]
  (*s) = (*s)[:len(*s) - 1]
  return d
} // end func Pop.
func (s *Stack) PrintStack() {
  for sx := 0; sx < len(*s); sx++ {
    fmt.Println("Type:", (*s)[sx].Mtype)
    structures.PrintMap((*s)[sx].Mmap, "Stacked map:")
    structures.PrintCoins((*s)[sx].Mcoins, "Stacked Coins")
  } // end for stack.
} // end func PrintStack.

func main() {

  skey := "QSoT0u9VlnrL4wq2pppy+jB4lEbYJ7xeWnE1VKzgVic="
  correspondent := "127.0.0.1"
  flag.Parse()
  flay := flag.Args()
  // Seed the random ramp.
  randramp = Seedramp(100)

  if len(flay) < 2 {
    fmt.Println("Usage: coin Command Configfile")
    os.Exit(1)
  } // endif flags.

  cmap := ParseFlags(flay, skey)
  structures.PrintMap(cmap, "Flags in main:")
  if cmap["cmd"] == "Error" {
    fmt.Printf(" '%s', # %s # \n", cmap["cmd"], cmap["strexp"])
    os.Exit(1)
  } // endif Error.

  // Individual commands, multiple transactions in a file or
  // automatically generated transactions.
  switch cmap["cmd"] {

    case "Multitest":
      cmdarray := ReadTests(cmap["scriptfile"])
      structures.Stringslice(cmdarray, "Multitest Commands")
      DoMultitest(cmap, cmdarray, skey, correspondent, time.Millisecond)

    case "Autotest":
      for ix:= 0; true; ix++ {
        cmdarray := Autoconstructor(cmap["throttle"], correspondent, cmap["pairs"], cmap["allpay"])
        structures.Stringslice(cmdarray, "Autotest Commands")
        DoMultitest(cmap, cmdarray, skey, correspondent, time.Millisecond)
         fmt.Printf("[%d] Wait 70s till Clearing is completed ... ", ix)
        time.Sleep(time.Second*80)
        fmt.Println("  ... Start next round.")
      } // end forever.

    default:
      coins, expl := SendMessages(cmap, correspondent)
      fmt.Println(expl)
      fmt.Printf("Balance reported: %d\n", structures.CoinValues(coins))

  } // end switch cmd.

} // end func main.


// Pre-populate the random ramp.
func Seedramp(orange int) []int {
  aramp := []int{}
  rand.Seed(int64(time.Now().UnixNano()))
  for ix := 0; ix < 10000; ix++ {
    nextrand := rand.Intn(orange)
    if nextrand == 0 { nextrand = 1 }
    aramp = append(aramp, nextrand)
  } // end for 10000.
  return aramp
} // end func Seedramp.



// Bundle Multitest commands from main in here.
func DoMultitest(dmap map[string]string, cmdarray []string, fedkey string, corresp string, delay time.Duration) {

  coins := []structures.Coin{}
  expl := ""

    for ix := 0; ix <len(cmdarray); ix++ {
      fmt.Printf("[%d] %s\n", ix, cmdarray[ix])
      argy := strings.Split(string(cmdarray[ix]), string(" "))
      mmap := ParseFlags(argy, fedkey)
      mmap["originalcmd"] = dmap["cmd"]
      coins, expl = SendMessages(mmap, corresp)
      fmt.Println(expl)
      structures.PrintCoins(coins, "OwnerOut")
      time.Sleep(delay)
    } // end for cmdarray.

} // end func DoMultitest.




func Autoconstructor(throt string, crosp string, pairs string, dir string) []string {
  cmds := []string{}
  utcmd := "Utxos conf/ubi.conf"
  cmds = append(cmds, utcmd)
  rand.Seed(int64(time.Now().Unix()))
  thrut, _ := strconv.Atoi(throt)

  for ix := 0; ix < thrut; ix++ {
    sender, receiver := ReturnPair(thrut, pairs, dir)
    payable := 0 // Randomized later in BuildPay.
    onecmd := fmt.Sprintf("PayCoins %s %d %s", sender, payable, receiver)
    cmds = append(cmds, onecmd)
  } // end for 120.

  return cmds

} // end func Autoconstructor.

func Randy(randels []int) int {
  return randels[rand.Intn(len(randels))]
} // end func Randy.

// ReturnPair: return a sender and receiver pair of names.
// Generalized to adjust length to actual number of users
func ReturnPair(userct int, path string, dir string) (string, string) {
  //fmt.Printf("----Row, Col : [%d-%d] ----\n", row,col)
  sen, rec, e := pairs.GetPair(path, row,col)
  if e != nil {
    methods.CheckError(e, true)
  }

  row += 1
  if row == userct {
    row = 0; col += 1
    if col == (userct) { col = 0 }
  }

  return dir+"/"+sen+".conf", dir+"/"+rec+".pub"
} // end func ReturnPair.


// Return true if element is in slice, false otherwise.
func Contains(inslice []int, element int) bool {

  for ix := 0; ix < len(inslice); ix++ {
    if inslice[ix] == element {
      return true
    } // endif inslice.
  } // end for inslice.

  return false

} // end func Contains.


// Given a config file, build the Balance command, send it and get the reply.
func GetBalance(clifile string, crosp string) int {
  amap := methods.GetConfigs(clifile)
  amap["cmd"] = "Balance"
  amap["who"] = clifile
  coins, _ := SendMessages(amap, crosp)
  // myhash := himitsu.HashPublicKey(amap["pubkey"])
  myhash := amap["pubhash"]
  tempbal := 0
  for ix := 0; ix < len(coins); ix++ {
    denom, _ := strconv.Atoi(coins[ix].Denom)
    tempbal = tempbal + denom
  } // end for coins.

  ballons[myhash] = tempbal
  baltime[myhash] = methods.MilliNow()
  allaccounts[myhash] = coins
  fmt.Printf("Balance for %s = %d.\n", clifile, tempbal)
  return tempbal
} // end func GetBalance.




// SendMessages: is the somewhat iterative replacement for SendAMessage.
func SendMessages(smap map[string]string, corresp string) ([]structures.Coin, string){
  mycoigns := []structures.Coin{}
  rx := structures.Transaction{}
  reply := ""
  newmess := Message{smap["cmd"], smap, mycoigns }
  stack.Put(newmess)
  fmt.Println("Stacked: ", stack.Peek())

  for (!stack.Empty()) {
    tosend := BuildMessage()
    if strings.Contains(tosend, "Error") {
      return mycoigns, smap["cmd"] + " Failed."
    } // endif error.

    reply = Transact(tosend, corresp, smap["cmd"])
    rx = UnwrapResult(reply)
    rerr := methods.VerifyTransaction(rx, false)
    methods.CheckErrorInst(0, rerr, true)
    if rerr != nil {
      fmt.Println("Verified? No.")
    }
    if rx.Ttyp == "Error" {
      fmt.Printf("%s: Transaction Failed.\n", rx.Ttyp)
    } // endif Error.
    stack.Pop()
    if !stack.Empty() {
      popper := stack.Pop()
      popper.Mcoins = rx.Outputs
      stack.Put(popper)
      // stack.PrintStack()
      // mash := himitsu.HashPublicKey(popper.Mmap["pubkey"])
      mash := popper.Mmap["pubhash"]
      ballons[mash] = structures.CoinValues(rx.Outputs)
      baltime[mash] = methods.MilliNow()
      allaccounts[mash] = rx.Outputs
    }
  } // end while non-empty stack

  reply = smap["cmd"] + " Transaction Completed."
  return rx.Outputs, reply

} // end func SendMessages.


// New BuildMessage: message is a Base64 string of a Balance, PayCoins or Quit
// command, constructed as a signed Transaction.
// Message to build comes from the stack.
func BuildMessage() string {

  stacktop := stack.Peek()

  switch(stacktop.Mtype) {
    case "Quit", "Utxos", "Balance", "Transactions":
      ans, msg := BuildClientCommand(stacktop.Mmap, stacktop.Mtype)
      return CheckUm(ans, msg)

    case "PayCoins":
      ans, msg := BuildPay(stacktop)
      return CheckUm(ans, msg)

    case "CreateCoins":
      ans, msg := BuildCreate(stacktop.Mmap, stacktop.Mcoins)
      return CheckUm(ans, msg)

    default:
      return "Error: Unrecognized Command."

  } // end switch cmd.

} // end func BuildMessage.



//BuildQuit: Put the Quit command into a Transaction.
func BuildQuit(qmap map[string]string) (string, string) {

  bx := newtxes.GetQuit(qmap)
  fmt.Println("Sending Quit for", qmap["who"])
  jbx,_ := json.Marshal(bx)
  btrx := base64.StdEncoding.EncodeToString(jbx)
  return btrx, "Good:"

} // end func BuildQuit.

//BuildBalance: Put the Balance command into a Transaction.
func BuildBalance(qmap map[string]string) (string, string) {

  bx := newtxes.GetBalance(qmap)
  fmt.Println("Sending GetBalance for", qmap["who"])
  jbx, _ := json.Marshal(bx)
  btrx := base64.StdEncoding.EncodeToString(jbx)
  return btrx, "Good:"

} // end func BuildBalance.

// BuildClientCommand:  put the request for Commandname into a str.Transaction
func BuildClientCommand(qmap map[string]string, commandname string) (string, string) {

  bx := newtxes.GetClientCommand(qmap, commandname)
  fmt.Printf("Sending Get%s for %s\n", commandname, qmap["who"])
  jbx, _ := json.Marshal(bx)
  btrx := base64.StdEncoding.EncodeToString(jbx)
  return btrx, "Good:"

} // end func BuildClientCommand.

// BuildPay: PayCoins might precipitate a Balance request, and possibly
// a Coin Compaction request. The original Paycoins must be stacked
// while these ancillary requests are serviced.
func BuildPay(top Message) (string, string) {

  // Map copy is needed twice below. Declare it here.
  mycoign := []structures.Coin{}
  pmap := make(map[string]string)
  for k, v := range top.Mmap {
    pmap[k] = v
  } // end copy map.

  // First see if there is a pre-fetched balance.
  getnew := true; getmine := 0
  // mash := himitsu.HashPublicKey(top.Mmap["pubkey"])
  mash := top.Mmap["pubhash"]

  if _, ok := ballons[mash]; ok {
    getnew = false
    timenow := methods.MilliNow()
    if timenow - baltime[mash] > int64(rollover) {
      getnew = true
    } // endif rollover.
  } // end if ballons.

  structures.IntMap(ballons, "SAVEDBALANCES")

  if getnew {
    // Stack a balance request on top of PayCoins.
   pmap["cmd"] = "Balance"
    newtop := Message{pmap["cmd"], pmap, mycoign}
    stack.Put(newtop)
    // fmt.Println("In BuildPay stacking Balance:", stack.Peek())
    return CheckUm(BuildBalance(pmap)), "Good"
  } else {
    getmine = ballons[mash]
    mycoign = allaccounts[mash]
  } // endif getnew.

  // If there is enough, do PayCoins, else bail out.
  // Randomization for Autotest deferred to here.  If pam
  // is 0, set it to random amount between zero and Balance.
  pam, _ := strconv.Atoi(top.Mmap["amount"])
  if pam == 0 {
    randone := Randy(randramp)
    randome := getmine
    pam = int(randome * randone / 100)
    if pam == 0 { pam = 1 }
    fmt.Printf("BuildPay: setting payment to %d out of %d\n", pam, randome)
  } // endif zero payments.

  if getmine < pam {
    return "Error", "Insufficient Funds: Payment Blocked."
  } else {
    fmt.Printf("My (%s) Balance = %d\n", top.Mmap["who"], getmine)
  } // endif low balance

  oneout := structures.Coin{}
  // Is there one coin large enough, or must we consolidate?
  structures.PrintCoins(allaccounts[mash], "BUILDPAYDIAGNOSTIC")
  if len(allaccounts[mash]) == 1 {
    oneout = allaccounts[mash][0]
    oneout.Owner = himitsu.BaseDER(top.Mmap["pubkey"])
    // structures.PrintCoin(oneout, "ONELARGECOIN")
  } else {
    // Is one of the many coins large enough?
    for sx := 0; sx < len(allaccounts[mash]); sx++ {
      denom, _ := strconv.Atoi(allaccounts[mash][sx].Denom)
      // fmt.Printf("OneLargeEnough? Matching %d and %d\n", denom, pam)
      // structures.PrintCoin(allaccounts[mash][sx], "BIGENOUGH")
      if denom > pam {
        oneout = allaccounts[mash][sx]
        // oneout.Owner = himitsu.BaseDER(top.Mmap["pubkey"])
        oneout.Owner = top.Mmap["pubder"]
        break
      } // endif denom.
    } // end for coins.
  } // endif supply

  // If no coin was large enough we do a Consolipay transaction.
  if oneout == (structures.Coin{}) {
    // fmt.Printf("BuildConsolipay: %s consolidates and pays %d to %s.\n", top.Mmap["who"], pam, top.Mmap["payee"])
    // transout := newtxes.Consolipay(top.Mmap, allaccounts[mash], pam, top.Mmap["payee"])
    fmt.Printf("BuildConsolipay: %s consolidates and pays %d to %s.\n", top.Mmap["who"], pam, top.Mmap["payhash"])
    transout := newtxes.Consolipay(top.Mmap, allaccounts[mash], pam, top.Mmap["payhash"])
    jtrx, _ := json.Marshal(transout)
    btrx := base64.StdEncoding.EncodeToString(jtrx)
    return btrx, "Good"
  } // endif noneout.

  // oneout is a good, big coin. make the payment.
  fmt.Printf("BuildPay: %s pays %d to %s\n", top.Mmap["who"], pam, top.Mmap["payee"])
  // trx := newtxes.PayCoins(top.Mmap, oneout, pam, top.Mmap["payee"])
  trx := newtxes.PayCoins(top.Mmap, oneout, pam, top.Mmap["payhash"])
  structures.PrintTransaction(trx, "IN build Pay")
  jtrx, _ := json.Marshal(trx)
  btrx := base64.StdEncoding.EncodeToString(jtrx)
  return btrx, "Good"

} // end func BuildPay.




// BuildCreate: only scrooge can build and send the CreateCoins transaction.
func BuildCreate(qmap map[string]string, nobal []structures.Coin) (string, string) {

  newcoy := newtxes.CreateCoins(qmap)
  coincount, _ := strconv.Atoi(qmap["coin"])
  fmt.Printf("Creating %d New Coins.\n\n", coincount)
  ctrx, _ := json.Marshal(newcoy)
  bctrx := base64.StdEncoding.EncodeToString(ctrx)
  return bctrx, "Good:"

} // end func BuildCreate.



//CheckUm: if msg contains Error, return it, else return ans.
func CheckUm(good string, bad string) string {
  if strings.Contains(bad, "Error:") {
    return bad
  } else {
    return good
  } // endif bad.

} // end func CheckUm.



// Implement 'element in array' method.
func CmdArrayHas(candidate string) bool {
  CmdArray := []string{ "Balance", "Transactions", "Quit", "Utxos", "PayCoins", "CreateCoins", "Multitest", "Autotest" }
    for _, element := range CmdArray {
        if element == candidate {
            return true
        }
    }
    return false

} // end func CmdArrayHas.


// The hash of the pubkey must match the privileged string.
func AuthorizedBy(pubfile string, privileged string) bool {

  inkey := himitsu.HashPublicKey(pubfile)
  return inkey == privileged

} // end func AuthorizedBy.



// Transact: handle the tcp connection setup, message send
// reply handle and close.
func Transact(msgin string, partner string, comd string) string {

  // Get connection handle.
  conn, err := net.Dial("tcp", partner + ":8081")
  methods.CheckErrorInst(1, err, true)

  // send to socket
  fmt.Fprintf(conn, msgin + "\n")

  // Quit: nothing coming back, pull the plug.
  if comd == "Quit" {
    fmt.Println("\nListener going away.")
    os.Exit(1)
  } // endif Quit.

  // listen for reply
  conn.SetReadDeadline(time.Now().Add(time.Second*10))
  message, errr := bufio.NewReader(conn).ReadString('\n')
  methods.CheckErrorInst(2, errr, true)

  // Process reply.
  // fmt.Print("Listener reply: " + message)
  time.Sleep(time.Millisecond)

  // Close the connection and  go away.
  conn.Close()

  return  message

} // end func Transact.


//UnwrapResult: Decode, Unmarshal and Print the reply.
func UnwrapResult(rez string) structures.Transaction {

  bx := structures.Transaction{}
  bsmsg, errb := base64.StdEncoding.DecodeString(rez)
  methods.CheckErrorInst(3, errb, true)
  err:= json.Unmarshal([]byte(bsmsg), &bx)
  methods.CheckErrorInst(4, err, true)
  // structures.PrintTransaction(bx, "\nReply Recvd:")
  return bx

} // end func UnwrapResult.


// ParseFlags: validate command and get configs.
func ParseFlags(flags []string, privilly string) map[string]string {

  cmap := make(map[string]string)
  cmd := "Error"; strexp := "OK."

  if CmdArrayHas(flags[0]) {
    cmd = flags[0]
  } else {
    cmd = "Error"
    strexp = "No such command: " + flags[0]
  } // endif cmd.

  if len(flags) > 1 {
    cmap = methods.GetConfigs(flags[1])
  } else {
    cmd = "Error"
    strexp = "Usage: ..."
  } // endif flags.

  switch(cmd) {

    case "Quit":
      if !AuthorizedBy(cmap["pubkey"], privilly) {
        cmd = "Error"
        strexp = "Unauthorized Quit instance."
      } // endif authorized.

    case "Utxos":
      strexp = "Update Balances in Listener."


    case "Balance":
      if methods.NoFile(flags[1]) || !strings.HasSuffix(flags[1], ".conf") {
        cmd = "Error"
        strexp = "Bad filename: " + flags[1]
      } // endif NoFile.

    case "Transactions":
      if methods.NoFile(flags[1]) || !strings.HasSuffix(flags[1], ".conf") {
        cmd = "Error"
        strexp = "Bad filename: " + flags[1]
      } // endif NoFile.

    case "Multitest":
      if methods.NoFile(flags[1]) {
        cmd = "Error"
        strexp = "Bad filename: " + flags[1]
      } else {
        cmap["scriptfile"] = flags[1]
      } // endif NoFile.

    case "Autotest":
      if methods.NoFile(flags[1]) {
        cmd = "Error"
        strexp = "Bad Directory Name: " + flags[1]
      } else {
        cmap["allpay"] = flags[1]
        cmap["throttle"] = flags[2]
        cmap["pairs"]=flags[3]
      } // endif NoFile.

    case "PayCoins":
      cmap["amount"] = flags[2]
      if methods.NoFile(flags[3]) {
        cmd = "Error"
        strexp = "Bad pubkey filename: " + flags[3]
      } else {
        cmap["payee"] = flags[3]
        cmap["payhash"] = himitsu.HashPublicKey(cmap["payee"])
      } // endif NoFile.

    case "CreateCoins":
      if !AuthorizedBy(cmap["pubkey"], privilly) {
        cmd = "Error"
        strexp = "Unauthorized Coin Creation."
      } // endif unauthorized.

      if methods.NoFile(flags[1]) {
        cmd = "Error"
        strexp = "Bad filename: " + flags[1]
      } // endif NoFile.

      // Coin specs are already in cmap["coin"] and cmap["denom"] so we're good.

    default:
      strexp = "Unknown command."
      cmd = "Error"

  } // end switch.

  cmap["cmd"] = cmd
  cmap["strexp"] = strexp
  return cmap

} // end func ParseFlags.

// Open a text file, read it, put results in an array.
func ReadTests(filein string) []string {
  slices := []string{}
  if file, err := os.Open(filein); err != nil {
    fmt.Println("Error:", err)
    os.Exit(0)
  } else {
    defer file.Close()
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
      slices = append(slices, scanner.Text())
    } // end for scanner.
  } // endif file.

  return slices

} // end func ReadTests.
