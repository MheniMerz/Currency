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
  March 17, 2017
*/


import (
  "flag"
  "fmt"
  "currency/himitsu"
  "currency/methods"
  "currency/structures"
  "os"
)

// All transactions for the given pub key.
func main() {

  flag.Parse()
  flay := flag.Args()

  if len(flay) == 0 {
    fmt.Println("Usage: doublespend block/ledger <users/name.pub>")
    os.Exit(1)
    } // endif flay.

  myledger := []structures.Transaction{}
  mycredits := map[string]int{}
  mypub := himitsu.BaseDER(flay[1])
  myledger = methods.FilterLedger(flay[0], myledger, mypub)

  for ix := 0; ix < len(myledger); ix++ {
    minput := myledger[ix].Inputs
    for jx := 0; jx < len(minput); jx++ {
      mykey := fmt.Sprintf("%013d/%s", minput[jx].Cid, minput[jx].Seq)
      if _, ok := mycredits[mykey]; ok {
        mycredits[mykey] += 1
      } else {
        mycredits[mykey] = 1
      } // endif coincount.
    } // end for Inputs.

  } // end for myledger.

  structures.MintMap(mycredits, "MyInputs")

} // end main.

