package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
	"trouvo/cs276parser"
	"trouvo/display"
	"trouvo/indexer"
	"trouvo/parser"
	// "trouvo/persist"
	"trouvo/search"
)

const (
	pathNameCACM       = "/Users/arthur/go/src/trouvo/Data/CACM/cacm.all"
	indexPathNameCACM  = "/Users/arthur/go/src/trouvo/Data/CACM/index.idx"
	pathNameCS276      = "/Users/arthur/go/src/trouvo/Data/CS276"
	indexPathNameCS276 = "/Users/arthur/go/src/trouvo/Data/CS276/index.idx"
	stopWordsPathName  = "/Users/arthur/go/src/trouvo/Data/CACM/common_words"
)

func main() {
	mainCS276()
}

func mainCACM() {
	indexer := buildCACM()
	runCACM(indexer)
}

func mainCS276() {
	indexers := buildCS276()
	runCS276(indexers)
}

func buildCACM() *indexer.Indexer {
	start := time.Now()
	p := parser.New(pathNameCACM, stopWordsPathName)
	p.Run() // Parsing...
	col := p.GetCollection()
	end := time.Now()
	elapsed := end.Sub(start)
	fmt.Println("Parsed in", elapsed.Round(time.Millisecond))
	fmt.Println("----")

	start = time.Now()
	docs := col.GetDocs()
	stopWords := p.GetStopWords()
	for _, doc := range docs {
		doc.Tokenize()
		doc.FilterTokens(stopWords)
	}
	end = time.Now()
	elapsed = end.Sub(start)
	fmt.Println("Tokenized in", elapsed.Round(time.Millisecond))
	fmt.Println("----")

	start = time.Now()
	col.BuildVocabulary()
	fmt.Println(len(col.GetVocabulary()), "words in vocabulary")
	end = time.Now()
	elapsed = end.Sub(start)
	fmt.Println("Vocabulary built in", elapsed.Round(time.Millisecond))
	fmt.Println("----")

	start = time.Now()
	indexer := indexer.New(col)
	indexer.Build()
	end = time.Now()
	elapsed = end.Sub(start)
	fmt.Println("Indexed in", elapsed.Round(time.Millisecond))
	fmt.Println("Index is", indexer.GetIndexSize(), "kB large.")
	fmt.Println("----")

	return indexer
}

func runCACM(indexer *indexer.Indexer) {
	engine := search.NewSearchEngine(
		indexer.GetIndex(),
		indexer.GetVocDict(),
		indexer.GetIdfDict(),
		indexer.GetDocDict(),
		indexer.GetDocNormDict(),
	)
	disp := display.New(indexer.GetDocDict())

	// Main infinite loop used to let the user query our search engine
	for true {
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("Type query :")
		text, _ := reader.ReadString('\n')
		start := time.Now()
		text = strings.TrimSpace(text)
		res := engine.VectSearch(text)
		end := time.Now()
		elapsed := end.Sub(start).Round(time.Microsecond)
		disp.Show(res, elapsed)
	}
}

func buildCS276() []*indexer.Indexer {
	start := time.Now()
	p := cs276parser.New(pathNameCS276)
	p.Run() // Parsing...
	end := time.Now()
	elapsed := end.Sub(start)
	fmt.Println("Parsed in", elapsed.Round(time.Second))
	fmt.Println("----")

	start = time.Now()
	cols := p.GetCollections()
	for _, col := range cols {
		col.BuildVocabulary()
	}
	end = time.Now()
	elapsed = end.Sub(start)
	fmt.Println("Vocabulary built in", elapsed.Round(time.Millisecond))
	fmt.Println("----")

	start = time.Now()
	indexers := []*indexer.Indexer{}
	indexSize := 0
	for _, col := range cols {
		indexer := indexer.New(col)
		indexer.Build()
		indexSize += indexer.GetIndexSize()
		indexers = append(indexers, indexer)
	}
	end = time.Now()
	elapsed = end.Sub(start)
	fmt.Println("Indexed in", elapsed.Round(time.Second))
	fmt.Println("Index is", indexSize, "kB large.")
	fmt.Println("----")
	return indexers
}

func runCS276(indexers []*indexer.Indexer) {
	// Build the engines from the indexers
	engines := []*search.Engine{}
	docDict := make(map[int]*parser.Document)
	for _, indexer := range indexers {
		engine := search.NewSearchEngine(
			indexer.GetIndex(),
			indexer.GetVocDict(),
			indexer.GetIdfDict(),
			indexer.GetDocDict(),
			indexer.GetDocNormDict(),
		)
		// Build the docDict from all the concatenated docDict
		// of each collection
		for docID, doc := range *indexer.GetDocDict() {
			docDict[docID] = doc
		}
		engines = append(engines, engine)
	}
	// Build the superEngine from all the different engines
	superEngine := search.NewSuperEngine(engines)
	disp := display.New(&docDict)

	// Main infinite loop used to let the user query our search engine
	for true {
		// Read the user query from the standard input
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("Type query :")
		text, _ := reader.ReadString('\n')
		start := time.Now()
		// Trim the useless spaces in the query
		text = strings.TrimSpace(text)
		// Run the query
		res := superEngine.Search(text)
		end := time.Now()
		elapsed := end.Sub(start).Round(time.Millisecond)
		// Display the results
		disp.Show(res, elapsed)
	}
}
