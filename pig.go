package main

import (
    "fmt"
    //"reflect"
    "os"
    "strconv"
    //"unicode"
    "math/rand"
    "regexp"
    "errors"
    "time"
)

// GLOBAL CONSTANTS
const PLAYER1 = "PLAYER1"
const PLAYER2 = "PLAYER2"
const DRAW = "DRAW"
const NO_OF_GAMES_IN_ONE_EVENT = 10
const SCORE_TO_END_A_GAME = 100

/* a single event is a set of 10 games played consecutively
 a game is a set of turns until one of the pig wins
 a series is a set of events(based on type of user input)
 
 if a user inputs a range(thresholds for which the strategy tests against)
 then system will simulate a series of events with each even having threshold as a number in given range against the competitor but both participants will have different threshold for testing strategies for sure

the cli arguments -
    there are two arguments
    each argument can be of two types 
        - a number ex: 23
        - range; ex: 23-40
    we access arguments using os.Args
*/

// turn -> game -> event -> series ( set of events )

// turn is a list of random integers in range 1 to 6(generated from rolling a die), and a score which is either 0 or something greater than threshold as per the strategy
type Turn struct {
    dieRollResults []int
    score int
}

// func to simulate rolling dice and return the results
func rollDice() int {
    rand.Seed(time.Now().UnixNano())
    result := rand.Intn(6) + 1
    return result
}

// method to simulate a player's turn
func (t *Turn) simulateTurn(threshold int) {
    // keep rolling a dice
    t.score = 0
    t.dieRollResults = []int{}
    for {
        rollResult := rollDice()
        t.dieRollResults = append(t.dieRollResults, rollResult)
        
        // if the dice rolls 1, turn is complete and score is reset to 0, it's next player's turn now
        if rollResult == 1 {
            t.score = 0
            return 
        } else {
            // if dice rolls greater than 1, than we add result to score
            t.score += rollResult
            // if the score is greater than threshold then we stop else we roll the dice again
            if t.score >= threshold {
                return
            }
        }
    }
}

// implementing Stringer interface on Turn
func (t Turn) String() string {
    return fmt.Sprintf("%v : %v\n", t.score, t.dieRollResults)
}

// game is a set of turns, totalScores of both players, the winner of the game, threshold 
type Game struct {
    // total scores for the game for both players
    p1Score, p2Score int
    // should winner field be declared as boolean isP1Winner
    winner string
    p1Turns, p2Turns []Turn
}

// implementing Stringer interface on Game 
func (g Game) String() string {
    return fmt.Sprintf("Total Score P1: %v, P2 : %v \nWinner: %v\n", g.p1Score, g.p2Score, g.winner) 
}

// method to simulate a game
// tsp1, tsp2 - threshold of strategy for player 1/2
func (g *Game) simulateGame(tsp1, tsp2 int) {
    g.p1Score = 0
    g.p2Score = 0
    g.p1Turns = []Turn{}
    g.p2Turns = []Turn{}

    // the players keep taking turns alternatively(consecutively in case of more than two players) until one of the players score crosses SCORE_TO_END_A_GAME
    for {
        // player 1's turn
        var p1Turn Turn
        p1Turn.simulateTurn(tsp1)
        g.p1Score += p1Turn.score
        g.p1Turns = append(g.p1Turns, p1Turn)

        // if p1's score is greater than or equal to SCORE_TO_END_A_GAME, than p1 is winner
        if g.p1Score >= SCORE_TO_END_A_GAME  {
            g.winner = PLAYER1
            return
        }

        var p2Turn = new(Turn)
        p2Turn.simulateTurn(tsp2)
        g.p2Score += p2Turn.score
        g.p2Turns = append(g.p2Turns, *p2Turn)

        // if p2's score is greater than or equal to 200, than p2 is winner
        if g.p2Score >= SCORE_TO_END_A_GAME  {
            g.winner = PLAYER2
            return
        }
    }
}

// an event is a set of 10 games, so we store number of wins for both players
type Event struct {
    games [NO_OF_GAMES_IN_ONE_EVENT]Game
    p1wins, p2wins int
    winner string
    // threshold as per the strategy for the entire game
    p1Threshold, p2Threshold int
}

// implementing Stringer interface on Turn
func (e Event) String() string {
    p1WinPercentage := (float64(e.p1wins*100)/float64(NO_OF_GAMES_IN_ONE_EVENT))
    p2WinPercentage := (float64(e.p2wins*100)/float64(NO_OF_GAMES_IN_ONE_EVENT))
    return fmt.Sprintf("Holding at %v VS Holding at %v: P1 wins %v/%v(%v%%) matches, P2 wins %v/%v(%v%%) matches\n", e.p1Threshold, e.p2Threshold, e.p1wins, NO_OF_GAMES_IN_ONE_EVENT, p1WinPercentage, e.p2wins, NO_OF_GAMES_IN_ONE_EVENT, p2WinPercentage)
}

// function to simulate an Event a set of 10 games
// tsp1, tsp2 - threshold of strategy for player 1/2

func (e *Event) simulateEvent(tsp1, tsp2 int) {
    e.games = [NO_OF_GAMES_IN_ONE_EVENT]Game{}
    e.p1wins = 0
    e.p2wins = 0
    e.p1Threshold = tsp1
    e.p2Threshold = tsp2

    // for each event, we simulate set of 10 games
    for i:=1; i<=10; i++ {
        var g Game
        g.simulateGame(tsp1, tsp2)
        e.games[i-1] = g
        if g.winner == PLAYER1 {
            e.p1wins += 1
        } else {
            e.p2wins += 1
        }
    }

    // based on the results, choose winner
    e.winner = getWinnerFromScores(e.p1wins, e.p2wins)
}

// series is a set of events along with results
type Series struct {
    p1Strategies, p2Strategies []int
    // map of threshold of strategy of player1 to all the events corresponding to that
    tsp1ToE map[int][]Event
    winner string
    p1Wins, p2Wins, drawMatches int
}

// simulate a series
func (s *Series)simulateSeries(p1Strategies, p2Strategies []int) {
    // series is a set of events
    // for each event, we pick unequal p1StrategyThreshold and p2StrategyThreshold from given parameters
    // and update s
    s.p1Strategies = p1Strategies
    s.p2Strategies = p2Strategies
    s.tsp1ToE = make(map[int][]Event)
    s.p1Wins = 0
    s.p2Wins = 0
    s.drawMatches = 0

    // looping through all strategy threshold of player1
    for _, pst1 := range p1Strategies {
        // having an event against all the strategy threshold of player2
        for _, pst2 := range p2Strategies {
            if pst1 != pst2 {
                // simulating the event with given strategy threshold
                var e Event
                e.simulateEvent(pst1, pst2)
                _, ok := s.tsp1ToE[pst1]
                if ok {
                   s.tsp1ToE[pst1] = append(s.tsp1ToE[pst1], e) 
                } else {
                    s.tsp1ToE[pst1] = []Event{e}
                }
                if e.winner == PLAYER1 {
                    s.p1Wins += 1
                } else if e.winner == PLAYER2 {
                    s.p2Wins += 1
                } else {
                    s.drawMatches += 1
                } 
            }
        }
        resultString, err := s.getEventResultTsp1(pst1)
        if err != nil {
            fmt.Printf("Error getting results for strategy having %d threshold\n")
        }
        fmt.Println(resultString)
    }

    // decide the winner
    s.winner = getWinnerFromScores(s.p1Wins, s.p2Wins)
}

// get winner from scores
func getWinnerFromScores(p1Score, p2Score int) string {
    var winner string
    if p1Score > p2Score {
        winner = PLAYER1
    } else if p1Score < p2Score {
        winner = PLAYER2
    } else {
        winner = DRAW
    }
    return winner
}

// func to check whether given string a number or not
func isNumber(s string) bool {
    /*
    // below method uses Atoi func of strconv package, identification is done based on error returned 
    _, err := strconv.Atoi(s)
    if err!=nil {
        fmt.Println("It's not a number, " + err.Error())
        //return false
    }
    //return true

    // method 2: using IsDigit func of unicode package
    for _, x := range s {
        if !unicode.IsDigit(x) {
            fmt.Printf("%v is not a number\n", x)
            // return false
        }
    }
    */
    
    // method 3: using regexp package, MustCompile func 
    digitCheck := regexp.MustCompile(`^[0-9]{1,3}$`)
    // fmt.Println(reflect.TypeOf(digitCheck.MatchString(s)))   // bool
    return digitCheck.MatchString(s)
}

// get number from string
func getNumberFromString(s string) (int, error) {
    n, err := strconv.Atoi(s)

    if err!=nil {
        fmt.Println("error : " + err.Error())
        return 0, errors.New(err.Error())
    }

    return n, nil
}

// func to get number from a string based on boolean flag, whether it's a number / range
func getThresholdsFromArg(s string, isThresholdNumber bool) []int {
    var nums = make([]int, 0, 2)
    if isThresholdNumber {
        n, err := getNumberFromString(s)
        if err!=nil {
            fmt.Println("error : "+ err.Error())
            return nums
        }
        nums = append(nums, n)
    } else {

        var dashIndex = -1
        for i, _ :=  range s {
            if s[i] == '-' {
                dashIndex = i 
                break
            }
        }
        n1, err := getNumberFromString(s[0: dashIndex])
        if err!=nil {
            fmt.Println("error : " + err.Error())
            return nums
        }
        n2, err := getNumberFromString(s[dashIndex+1:])
        if err!=nil {
            fmt.Println("error : " + err.Error())
            return nums
        }
        nums = append(nums, n1)
        nums = append(nums, n2)
    }
    return nums
}

func main() {
    // Record the start time
    startTime := time.Now()

    fmt.Println("Welcome to the pig game!")
    //fmt.Println(reflect.TypeOf(os.Args))    // []string
    //fmt.Println(os.Args)                    // [./pig 12 12-34]

    if len(os.Args) < 3 {
        fmt.Println("Please provide threshold of strategy for the game")
        return
    }
    isStrategy1Num := isNumber(os.Args[1]) 
    isStrategy2Num := isNumber(os.Args[2]) 

    // thresholds for both players
    t1 := getThresholdsFromArg(os.Args[1], isStrategy1Num)
    t2 := getThresholdsFromArg(os.Args[2], isStrategy2Num)

    // check if the argument is a number / range
    if(isStrategy1Num && isStrategy2Num) {

        // story 1
        // simulate an event 
        var g = new(Event)
        g.simulateEvent(t1[0], t2[0])
        fmt.Println(*g)

    } else if(isStrategy1Num && !isStrategy2Num) {
        
        // story 2
        var s = new(Series)
        p2Strategies := make([]int, t2[1]-t2[0]+1, t2[1]-t2[0]+1)
        for i,j := t2[0],0; i<=t2[1]; i++{
            p2Strategies[j] = i
            j++
        }
        fmt.Println("Simulation has started, waiting for the results...")
        s.simulateSeries(t1, p2Strategies)
        // fmt.Println(*s)
        // show results
        for _, v := range s.tsp1ToE[t1[0]] {
           fmt.Println(v) 
        }
        
        fmt.Println("Summary:\nWinner of the series is : " + s.winner)
        fmt.Printf("Matches won by Player 1 : %d/%d\n" , s.p1Wins, len(s.tsp1ToE[t1[0]]))
        fmt.Printf("Matches won by Player 2 : %d/%d\n" , s.p2Wins, len(s.tsp1ToE[t1[0]]))
        fmt.Printf("Matches Draw: %d/%d\n" , s.drawMatches, len(s.tsp1ToE[t1[0]]))

    } else if(!isStrategy1Num && !isStrategy2Num) {
        
        // story 3
        var s = new(Series)
        p1Strategies := make([]int, t1[1]-t1[0]+1, t1[1]-t1[0]+1)
        for i,j := t1[0],0; i<=t1[1]; i++{
            p1Strategies[j] = i
            j++
        }
        p2Strategies := make([]int, t2[1]-t2[0]+1, t2[1]-t2[0]+1)
        for i,j := t2[0],0; i<=t2[1]; i++{
            p2Strategies[j] = i
            j++
        }
        fmt.Println("Simulation has started, waiting for the results...")
        // simulate along with showing result for each event as in when it completes
        s.simulateSeries(p1Strategies, p2Strategies)
        
    }
    // Record the end time
    endTime := time.Now()

    // Calculate the duration
    duration := endTime.Sub(startTime)

    // Print the duration in seconds and milliseconds
    fmt.Printf("Time taken: %v\n", duration)
}

/* func to display result as required in story 3
   $ ./pig 1-100 1-100
    Result: Wins, losses staying at k =   1: 277/990 (28.0%), 713/990 (72.0%)
    Result: Wins, losses staying at k =   2: 237/990 (23.9%), 753/990 (76.1%)
    Result: Wins, losses staying at k =   3: 322/990 (32.5%), 668/990 (67.5%)
    ...
    ...
    Result: Wins, losses staying at k =  99: 225/990 (22.7%), 765/990 (77.3%)
    Result: Wins, losses staying at k = 100: 210/990 (21.2%), 780/990 (78.8%)
*/
func (s *Series)getEventResultTsp1(tsp1 int) (string, error) {
    events, ok := s.tsp1ToE[tsp1]
    var resultString = ""
    if !ok {
        return resultString, errors.New("There are no events corresponding to strategy with threshold " + strconv.Itoa(tsp1))
    }
    // loop over all the events and check how many games were won by each player
    p1Wins, p2Wins, draws := 0,0,0
    totalGames := len(events)*10
    for _, e := range events {
        for _, game := range e.games {
            if game.winner == PLAYER1 {
                p1Wins += 1
            } else if game.winner == PLAYER2 {
                p2Wins += 1
            } else {
                draws += 1
            }
        }
    }
    p1WinPercentage := (float64(p1Wins*100)/float64(totalGames))
    p2WinPercentage := (float64(p2Wins*100)/float64(totalGames))
    resultString = fmt.Sprintf("Result: Wins, losses staying at k = %d : %v/%v(%.2f%%) , %v/%v(%.2f%%) \n", tsp1, p1Wins, totalGames, p1WinPercentage, p2Wins, totalGames, p2WinPercentage) 
    return resultString, nil
}
