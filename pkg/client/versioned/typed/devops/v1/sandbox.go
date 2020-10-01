// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	"context"
	"time"

	v1 "gitlab.com/stackvista/devops/devopserator/apis/devops/v1"
	scheme "gitlab.com/stackvista/devops/devopserator/pkg/client/versioned/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// SandboxesGetter has a method to return a SandboxInterface.
// A group's client should implement this interface.
type SandboxesGetter interface {
	Sandboxes() SandboxInterface
}

// SandboxInterface has methods to work with Sandbox resources.
type SandboxInterface interface {
	Create(ctx context.Context, sandbox *v1.Sandbox, opts metav1.CreateOptions) (*v1.Sandbox, error)
	Update(ctx context.Context, sandbox *v1.Sandbox, opts metav1.UpdateOptions) (*v1.Sandbox, error)
	UpdateStatus(ctx context.Context, sandbox *v1.Sandbox, opts metav1.UpdateOptions) (*v1.Sandbox, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.Sandbox, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1.SandboxList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.Sandbox, err error)
	SandboxExpansion
}

// sandboxes implements SandboxInterface
type sandboxes struct {
	client rest.Interface
}

// newSandboxes returns a Sandboxes
func newSandboxes(c *DevopsV1Client) *sandboxes {
	return &sandboxes{
		client: c.RESTClient(),
	}
}

// Get takes name of the sandbox, and returns the corresponding sandbox object, and an error if there is any.
func (c *sandboxes) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.Sandbox, err error) {
	result = &v1.Sandbox{}
	err = c.client.Get().
		Resource("sandboxes").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Sandboxes that match those selectors.
func (c *sandboxes) List(ctx context.Context, opts metav1.ListOptions) (result *v1.SandboxList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1.SandboxList{}
	err = c.client.Get().
		Resource("sandboxes").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested sandboxes.
func (c *sandboxes) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Resource("sandboxes").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a sandbox and creates it.  Returns the server's representation of the sandbox, and an error, if there is any.
func (c *sandboxes) Create(ctx context.Context, sandbox *v1.Sandbox, opts metav1.CreateOptions) (result *v1.Sandbox, err error) {
	result = &v1.Sandbox{}
	err = c.client.Post().
		Resource("sandboxes").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(sandbox).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a sandbox and updates it. Returns the server's representation of the sandbox, and an error, if there is any.
func (c *sandboxes) Update(ctx context.Context, sandbox *v1.Sandbox, opts metav1.UpdateOptions) (result *v1.Sandbox, err error) {
	result = &v1.Sandbox{}
	err = c.client.Put().
		Resource("sandboxes").
		Name(sandbox.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(sandbox).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *sandboxes) UpdateStatus(ctx context.Context, sandbox *v1.Sandbox, opts metav1.UpdateOptions) (result *v1.Sandbox, err error) {
	result = &v1.Sandbox{}
	err = c.client.Put().
		Resource("sandboxes").
		Name(sandbox.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(sandbox).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the sandbox and deletes it. Returns an error if one occurs.
func (c *sandboxes) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.client.Delete().
		Resource("sandboxes").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *sandboxes) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Resource("sandboxes").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched sandbox.
func (c *sandboxes) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.Sandbox, err error) {
	result = &v1.Sandbox{}
	err = c.client.Patch(pt).
		Resource("sandboxes").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
