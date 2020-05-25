package git

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/ssrathi/gogit/util"
)

var (
	repoDir    string
	repo       *Repo
	blob       *Blob
	tree       *Tree
	commit     *Commit
	testFile   string
	blobHash   string
	treeHash   string
	commitHash string
	testData   string
	treeInput  string
	commitMsg  string
)

// assertEqual checks if two given values are equal and fatals if not.
func assertEqual(t *testing.T, got interface{}, want interface{}) {
	t.Helper()

	if got != want {
		t.Fatalf("got '%+v' (%v), want '%+v' (%v)", got, reflect.TypeOf(got),
			want, reflect.TypeOf(want))
	}
}

// setupTestArtifacts creates objects in a new repo for testing gogit commands.
func setupTestArtifacts(t *testing.T) {
	t.Helper()

	// Disable internal logs during test runs.
	log.SetOutput(ioutil.Discard)
	log.SetFlags(0)

	var err error
	repoDir, err = ioutil.TempDir(os.TempDir(), "testGoGit")
	assertEqual(t, err, nil)
	t.Logf("Repository directory: %s", repoDir)

	repo, err = NewRepo(repoDir)
	assertEqual(t, err, nil)
	testFile = "testfile"
	tmpFile, err := os.OpenFile(filepath.Join(repoDir, testFile),
		os.O_WRONLY|os.O_CREATE, 0644)
	assertEqual(t, err, nil)

	testData = "Hello World\n"
	_, err = tmpFile.WriteString(testData)
	assertEqual(t, err, nil)
	tmpFile.Close()

	// Save this file's data as a blob first.
	blob, err = NewBlobFromFile(repo, tmpFile.Name())
	assertEqual(t, err, nil)
	blobHash, err = repo.ObjectWrite(blob.Object, true)
	assertEqual(t, err, nil)

	// Create a tree with this blob
	treeInput = fmt.Sprintf("100644 blob %s\t%s\n",
		blobHash, filepath.Base(tmpFile.Name()))
	t.Logf("Tree input: %s", treeInput)
	tree, err = NewTreeFromInput(repo, treeInput)
	assertEqual(t, err, nil)

	// Write the tree now.
	treeHash, err = repo.ObjectWrite(tree.Object, true)
	assertEqual(t, err, nil)

	// Make a commit with this tree.
	commitMsg = "Test commit for testing"
	commit, err = NewCommitFromParams(repo, treeHash, "", commitMsg)
	assertEqual(t, err, nil)

	// Write the commit now.
	commitHash, err = repo.ObjectWrite(commit.Object, true)
	assertEqual(t, err, nil)

	t.Logf("Blob  : %s", blobHash)
	t.Logf("Tree  : %s", treeHash)
	t.Logf("Commit: %s", commitHash)

	// Write the commit hash to master branch reference manually.
	// TODO: Use internal API once 'update-ref' is implemented.
	masterFile, _ := repo.FilePath(false, "refs", "heads", "master")
	err = ioutil.WriteFile(masterFile, []byte(commitHash+"\n"), 0644)
	assertEqual(t, err, nil)
}

func TestCommands(t *testing.T) {
	setupTestArtifacts(t)
	defer os.RemoveAll(repoDir)

	// Validate 'gogit' init operations by checking various folders and files
	// in the repo.
	t.Run("Validate repository", func(t *testing.T) {
		gitDir := filepath.Join(repoDir, ".git")
		assertEqual(t, util.IsPathPresent(gitDir), true)
		assertEqual(t, util.IsPathPresent(filepath.Join(gitDir, "HEAD")), true)
		assertEqual(t, util.IsPathPresent(filepath.Join(gitDir, "description")), true)
		assertEqual(t, util.IsPathPresent(filepath.Join(gitDir, "config")), true)
		assertEqual(t, util.IsPathPresent(filepath.Join(gitDir, "objects")), true)
		assertEqual(t, util.IsPathPresent(filepath.Join(gitDir, "refs")), true)
		assertEqual(t, util.IsPathPresent(filepath.Join(gitDir, "refs", "tags")), true)
		assertEqual(t, util.IsPathPresent(filepath.Join(gitDir, "refs", "heads")), true)
	})

	// Validate that a repo creation in a non-emptry dir fails.
	t.Run("Validate non-empty repository creation failure", func(t *testing.T) {
		_, err := NewRepo(repoDir)
		want := fmt.Sprintf("Work-tree %q is not empty", repoDir)
		assertEqual(t, err.Error(), want)
	})

	// Validate the blob hash from hash-object operation.
	// "want" is the output of 'echo "Hello World" | git hash-object --stdin'
	t.Run("Validate blob hash", func(t *testing.T) {
		got := blobHash
		want := "557db03de997c86a4a028e1ebd3a1ceb225be238"
		assertEqual(t, got, want)
	})

	// Validate the tree hash from mktree operation.
	// "want" is the output of 'git mktree' with input
	// '100644 blob 557db03de997c86a4a028e1ebd3a1ceb225be238	testfile'
	t.Run("Validate tree hash", func(t *testing.T) {
		got := treeHash
		want := "e592dfe791dd1e1cf202668707a5cfac07a635b3"
		assertEqual(t, got, want)
	})

	// Validate various cat-file options on a blob object.
	t.Run("Validate cat-file blob -p option", func(t *testing.T) {
		got := blob.Print()
		want := testData
		assertEqual(t, got, want)
	})

	t.Run("Validate cat-file blob -t option", func(t *testing.T) {
		got := blob.Type()
		want := "blob"
		assertEqual(t, got, want)
	})

	t.Run("Validate cat-file blob -s option", func(t *testing.T) {
		got := blob.DataSize()
		want := len(testData)
		assertEqual(t, got, want)
	})

	// Validate various cat-file options on a tree object.
	t.Run("Validate cat-file tree -p option", func(t *testing.T) {
		got := tree.Print()
		want := treeInput
		assertEqual(t, got, want)
	})

	t.Run("Validate cat-file tree -t option", func(t *testing.T) {
		got := tree.Type()
		want := "tree"
		assertEqual(t, got, want)
	})

	t.Run("Validate cat-file tree -s option", func(t *testing.T) {
		got := tree.DataSize()
		// "git cat-file -s" for this gives 36. This is not the same as lenght
		// of "treeInput" as the blob hash is stored as binary bytes.
		want := 36
		assertEqual(t, got, want)
	})

	// Validate various cat-file options on a commit object.
	t.Run("Validate cat-file commit -t option", func(t *testing.T) {
		got := commit.Type()
		want := "commit"
		assertEqual(t, got, want)
	})

	// Validate that the tree inside the commit matches the given tree hash.
	t.Run("Validate tree-hash inside commit", func(t *testing.T) {
		got := commit.TreeHash()
		want := treeHash
		assertEqual(t, got, want)
	})

	// Validate that the msg inside the commit matches the given message.
	t.Run("Validate message inside commit", func(t *testing.T) {
		got := commit.Msg
		want := commitMsg
		assertEqual(t, got, want)
	})

	// Validate that the parent inside the commit matches the given parent.
	t.Run("Validate parent inside commit", func(t *testing.T) {
		parents := commit.Parents()
		// No parent is given to a base commit.
		assertEqual(t, len(parents), 0)
	})

	// Validate that the author details inside the commit matches the given values.
	t.Run("Validate author inside commit", func(t *testing.T) {
		name, email := commit.Author()
		assertEqual(t, name, AuthorName)
		assertEqual(t, email, AuthorEmail)
	})

	// Validate that a blob can be parsed from a given blob-hash.
	t.Run("Validate blob creation with a blob-hash", func(t *testing.T) {
		obj, err := repo.ObjectParse(blobHash)
		assertEqual(t, err, nil)

		testBlob, err := NewBlob(repo, obj)
		assertEqual(t, err, nil)

		assertEqual(t, testBlob.Print(), testData)
		assertEqual(t, testBlob.DataSize(), len(testData))
	})

	// Validate that a tree can be parsed from a given tree-hash.
	t.Run("Validate tree creation with a tree-hash", func(t *testing.T) {
		obj, err := repo.ObjectParse(treeHash)
		assertEqual(t, err, nil)

		testTree, err := NewTree(repo, obj)
		assertEqual(t, err, nil)

		// "git cat-file -s" for this gives 36. This is not the same as lenght
		// of "treeInput" as the blob hash is stored as binary bytes.
		assertEqual(t, testTree.Print(), treeInput)
		assertEqual(t, testTree.DataSize(), 36)
	})

	// Validate that a commit can be parsed from a given commit-hash.
	t.Run("Validate commit creation with a commit-hash", func(t *testing.T) {
		obj, err := repo.ObjectParse(commitHash)
		assertEqual(t, err, nil)

		testCommit, err := NewCommit(repo, obj)
		assertEqual(t, err, nil)

		assertEqual(t, testCommit.Msg, commitMsg)
		assertEqual(t, testCommit.TreeHash(), treeHash)

		name, email := testCommit.Author()
		assertEqual(t, name, AuthorName)
		assertEqual(t, email, AuthorEmail)
	})

	// Validate various rev-parse arguments.
	t.Run("Validate rev-parse HEAD", func(t *testing.T) {
		objHash, err := repo.UniqueNameResolve("HEAD")
		assertEqual(t, err, nil)
		assertEqual(t, objHash, commitHash)
	})

	t.Run("Validate rev-parse short hash", func(t *testing.T) {
		objHash, err := repo.UniqueNameResolve(commitHash[:4])
		assertEqual(t, err, nil)
		assertEqual(t, objHash, commitHash)
	})

	t.Run("Validate rev-parse master", func(t *testing.T) {
		objHash, err := repo.UniqueNameResolve("master")
		assertEqual(t, err, nil)
		assertEqual(t, objHash, commitHash)
	})

	t.Run("Validate rev-parse heads/master", func(t *testing.T) {
		objHash, err := repo.UniqueNameResolve("heads/master")
		assertEqual(t, err, nil)
		assertEqual(t, objHash, commitHash)
	})

	t.Run("Validate rev-parse refs/heads/master", func(t *testing.T) {
		objHash, err := repo.UniqueNameResolve("refs/heads/master")
		assertEqual(t, err, nil)
		assertEqual(t, objHash, commitHash)
	})

	t.Run("Validate rev-parse less than four short hash", func(t *testing.T) {
		_, err := repo.UniqueNameResolve(commitHash[:3])
		want := fmt.Sprintf("fatal: ambiguous argument '%s': unknown revision "+
			"or path not in the working tree", commitHash[:3])
		assertEqual(t, err.Error(), want)
	})

	t.Run("Validate rev-parse invalid", func(t *testing.T) {
		_, err := repo.UniqueNameResolve("FOO")
		want := fmt.Sprintf("fatal: ambiguous argument 'FOO': unknown revision " +
			"or path not in the working tree")
		assertEqual(t, err.Error(), want)
	})

	// Validate 'show-ref' outputs.
	t.Run("Validate show-ref", func(t *testing.T) {
		refs, err := repo.GetRefs("", false /* showHead */)
		assertEqual(t, err, nil)
		assertEqual(t, len(refs), 1)
		assertEqual(t, refs[0].RefHash, commitHash)
		assertEqual(t, refs[0].Name, "refs/heads/master")
	})

	t.Run("Validate show-ref with HEAD", func(t *testing.T) {
		refs, err := repo.GetRefs("", true /* showHead */)
		assertEqual(t, err, nil)
		assertEqual(t, len(refs), 2)

		assertEqual(t, refs[0].RefHash, commitHash)
		assertEqual(t, refs[0].Name, "HEAD")

		assertEqual(t, refs[1].RefHash, commitHash)
		assertEqual(t, refs[1].Name, "refs/heads/master")
	})

	// Validate 'checkout' functionality.
	t.Run("Validate checkout", func(t *testing.T) {
		checkoutDir, err := ioutil.TempDir(os.TempDir(), "testGoGitCheckout")
		assertEqual(t, err, nil)
		defer os.RemoveAll(checkoutDir)

		t.Logf("Checkout test directory: %s", checkoutDir)
		err = tree.Checkout(checkoutDir)
		assertEqual(t, err, nil)

		// testFile should be created in this new path with original data.
		dataFile := filepath.Join(checkoutDir, testFile)
		assertEqual(t, util.IsPathPresent(dataFile), true)
		data, err := ioutil.ReadFile(dataFile)
		assertEqual(t, err, nil)
		assertEqual(t, string(data), testData)
	})
}