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
  March 12, 2017
*/


import (
  "flag"
  "fmt"
  "os"
  "currency/himitsu"
  "currency/methods"
  "currency/structures"
  "strconv"
)

// Filter all coins out of the ledger.
func allcoins(userdir string, hgt int) {

  ledger := []structures.Transaction{}
  ledger = methods.LoadLedger("blocks/ledger", ledger)
  if hgt > 0 {
    ledger = ledger[:hgt]
  }
  m1 := methods.M1(ledger)
  allpubs := methods.GetDir(userdir, ".pub")

  for fx := 0; fx < len(allpubs); fx++ {
    onepub := himitsu.HashPublicKey(allpubs[fx])
    ones := methods.GetMyCoins(m1, onepub)
    structures.PrintCoins(ones, allpubs[fx])
    fmt.Printf("Balance for %s is %d\n", allpubs[fx], structures.CoinCount(ones))
  } // end for allusers.

} // end func allcoins.

// Get the (pubkey) args and run allcoins.
func main() {

  ht := 0
  flag.Parse()
  flay := flag.Args()
  if len(flay) == 0 {
    fmt.Println("Usage: allcoins <userdir>" [ht])
    os.Exit(1)
  } // end if flay.

  if len(flay) == 2 {
    ht, _ = strconv.Atoi(flay[1])
  } else { ht = 0 }

  allcoins(flay[0], ht)

} // end main.



