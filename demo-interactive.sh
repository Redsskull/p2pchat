#!/bin/bash

# P2P Chat Interactive Demo Script
# Demonstrates the new user-friendly CLI experience

echo "🌟 P2P Chat Interactive CLI Demo"
echo "================================="
echo ""
echo "This demo shows how easy it is to start P2P Chat with the new interactive mode!"
echo ""

# Check if binary exists
if [ ! -f "./p2pchat" ]; then
    echo "❌ p2pchat binary not found. Please build it first:"
    echo "   go build -o p2pchat cmd/p2pchat/main.go"
    exit 1
fi

echo "📋 Available demo modes:"
echo "   1) Interactive mode (prompts for username)"
echo "   2) Quick start with username"
echo "   3) Manual configuration"
echo "   4) Debug mode"
echo "   5) Help message"
echo "   6) Multi-user simulation"
echo ""

read -p "Choose demo mode (1-6): " choice

case $choice in
    1)
        echo ""
        echo "🎯 Demo 1: Interactive Mode"
        echo "=========================="
        echo ""
        echo "Simply run './p2pchat' and you'll be prompted for your username."
        echo "The port will be automatically assigned!"
        echo ""
        echo "Try it now:"
        echo "./p2pchat"
        echo ""
        echo "Press Ctrl+C when you're done exploring the interface."
        ./p2pchat
        ;;

    2)
        echo ""
        echo "🎯 Demo 2: Quick Start with Username"
        echo "===================================="
        echo ""
        echo "Specify username directly, port auto-assigned:"
        echo "./p2pchat -username demo_user"
        echo ""
        echo "Press Ctrl+C when you're done exploring the interface."
        ./p2pchat -username demo_user
        ;;

    3)
        echo ""
        echo "🎯 Demo 3: Manual Configuration"
        echo "==============================="
        echo ""
        echo "Full control over username and port:"
        echo "./p2pchat -username alice -port 8080"
        echo ""
        echo "Press Ctrl+C when you're done exploring the interface."
        ./p2pchat -username alice -port 8080
        ;;

    4)
        echo ""
        echo "🎯 Demo 4: Debug Mode"
        echo "===================="
        echo ""
        echo "Interactive mode with debug logging enabled:"
        echo "./p2pchat -debug"
        echo ""
        echo "Debug information will be logged to 'p2pchat-debug.log'"
        echo "Press Ctrl+C when you're done exploring the interface."
        ./p2pchat -debug
        ;;

    5)
        echo ""
        echo "🎯 Demo 5: Help Message"
        echo "======================="
        echo ""
        ./p2pchat -help
        ;;

    6)
        echo ""
        echo "🎯 Demo 6: Multi-User Simulation"
        echo "================================"
        echo ""
        echo "This will show how multiple users can start simultaneously"
        echo "with automatic port assignment preventing conflicts."
        echo ""
        echo "We'll start 3 users in background (they'll exit quickly since no TTY):"
        echo ""

        echo "Starting alice..."
        echo "alice" | timeout 3s ./p2pchat &
        sleep 1

        echo "Starting bob..."
        echo "bob" | timeout 3s ./p2pchat &
        sleep 1

        echo "Starting charlie..."
        echo "charlie" | timeout 3s ./p2pchat &
        sleep 1

        echo ""
        echo "Notice how each user gets a different port automatically:"
        echo "- alice: likely got port 8080"
        echo "- bob: likely got port 8081"
        echo "- charlie: likely got port 8082"
        echo ""
        echo "This prevents port conflicts when multiple users start simultaneously!"

        # Wait for background processes
        wait
        ;;

    *)
        echo "❌ Invalid choice. Please run the script again and choose 1-6."
        exit 1
        ;;
esac

echo ""
echo "🎉 Demo complete! Key improvements:"
echo "   ✅ No more required command line arguments"
echo "   ✅ Interactive username prompt"
echo "   ✅ Automatic port assignment (8080-8999 range)"
echo "   ✅ Collision detection for multiple users"
echo "   ✅ Helpful startup messages"
echo "   ✅ Command line options still work for power users"
echo ""
echo "📖 For more info: ./p2pchat -help"
echo "🚀 Ready for production use!"
