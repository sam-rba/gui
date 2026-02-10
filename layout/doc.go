/*
Package layout provides means of partitioning screen space.

A layout exists in a parent Env, and has one or more child Envs. It
acts as a multiplexer for the children. The parent Env may be a window,
another layout, or whatever... Several layers of layouts can be composed.

A layout allocates screen area to its children by intercepting Resize
events from the parent Env. Upon reception by the layout, a Resize event
is transformed for each child, and forwarded to them.

Draw calls from the children are intercepted and translated onto their
respective areas before being forwarded to the parent Env.
*/
package layout
