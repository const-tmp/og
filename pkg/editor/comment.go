package editor

import "go/ast"

// Comment4Node returns *ast.CommentGroup with text positioned before node
func Comment4Node(node ast.Node, text string) *ast.CommentGroup {
	return &ast.CommentGroup{List: []*ast.Comment{{Text: text, Slash: node.Pos() - 1}}}
}
