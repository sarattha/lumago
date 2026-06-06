package desktop

/*
#cgo CFLAGS: -I/opt/homebrew/include
#include <dlfcn.h>
#include <stdint.h>
#define GLFW_INCLUDE_VULKAN
#include <GLFW/glfw3.h>
#include <vulkan/vulkan.h>

static inline int lumagoPreloadVulkanLoader(void) {
	void* handle = dlopen("/opt/homebrew/lib/libvulkan.1.dylib", RTLD_LAZY | RTLD_GLOBAL);
	return handle != NULL;
}

static inline VkResult lumagoCreateWindowSurface(VkInstance instance, GLFWwindow* window, VkSurfaceKHR* surface) {
	return glfwCreateWindowSurface(instance, window, NULL, surface);
}
*/
import "C"

import (
	"errors"
	"os"
	"runtime"
	"unsafe"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/sarattha/lumago/engine/input"
)

type Window struct {
	window *glfw.Window
}

func NewWindow(width, height int, title string) (*Window, error) {
	runtime.LockOSThread()
	seedLoaderPath("DYLD_LIBRARY_PATH", "/opt/homebrew/lib")
	seedLoaderPath("DYLD_FALLBACK_LIBRARY_PATH", "/opt/homebrew/lib")
	C.lumagoPreloadVulkanLoader()

	if err := glfw.Init(); err != nil {
		return nil, err
	}
	if !glfw.VulkanSupported() {
		glfw.Terminate()
		return nil, errors.New("desktop: Vulkan loader is not available to GLFW")
	}

	glfw.DefaultWindowHints()
	glfw.WindowHint(glfw.ClientAPI, glfw.NoAPI)
	glfw.WindowHint(glfw.Resizable, glfw.True)

	window, err := glfw.CreateWindow(width, height, title, nil, nil)
	if err != nil {
		glfw.Terminate()
		return nil, err
	}

	return &Window{window: window}, nil
}

func (w *Window) RequiredInstanceExtensions() []string {
	return w.window.GetRequiredInstanceExtensions()
}

func (w *Window) CreateSurface(instance unsafe.Pointer) (unsafe.Pointer, error) {
	window := *(**C.GLFWwindow)(unsafe.Pointer(w.window))
	var surface C.VkSurfaceKHR
	result := C.lumagoCreateWindowSurface(C.VkInstance(instance), window, &surface)
	if result != C.VK_SUCCESS {
		return nil, errors.New("desktop: failed to create Vulkan window surface")
	}
	return unsafe.Pointer(surface), nil
}

func (w *Window) ShouldClose() bool {
	return w.window.ShouldClose()
}

func (w *Window) PollEvents() {
	glfw.PollEvents()
}

func (w *Window) FramebufferSize() (int, int) {
	return w.window.GetFramebufferSize()
}

func (w *Window) WaitForFramebuffer() {
	for {
		if w.ShouldClose() {
			return
		}
		width, height := w.FramebufferSize()
		if width > 0 && height > 0 {
			return
		}
		glfw.WaitEvents()
	}
}

func (w *Window) KeyDown(key input.Key) bool {
	if w.window == nil {
		return false
	}
	return w.window.GetKey(glfwKey(key)) == glfw.Press
}

func (w *Window) Close() {
	if w.window != nil {
		w.window.Destroy()
		w.window = nil
	}
	glfw.Terminate()
}

func glfwKey(key input.Key) glfw.Key {
	switch key {
	case input.KeySpace:
		return glfw.KeySpace
	case input.KeyUp:
		return glfw.KeyUp
	case input.KeyDown:
		return glfw.KeyDown
	case input.KeyW:
		return glfw.KeyW
	case input.KeyS:
		return glfw.KeyS
	case input.KeyR:
		return glfw.KeyR
	case input.KeyEscape:
		return glfw.KeyEscape
	default:
		return glfw.KeyUnknown
	}
}

func VulkanProcAddr() unsafe.Pointer {
	return glfw.GetVulkanGetInstanceProcAddress()
}

func seedLoaderPath(name, path string) {
	value := os.Getenv(name)
	if value == "" {
		_ = os.Setenv(name, path)
		return
	}
	if value == path {
		return
	}
	_ = os.Setenv(name, path+":"+value)
}
