# miniServer-completion.bash

_miniServer() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    opts="-d -k -kl -l -p -v"

    if [[ ${cur} == -* ]] ; then
        COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
        return 0
    fi

    case "${prev}" in
        -d)
            COMPREPLY=( $(compgen -d -- ${cur}) )
            return 0
            ;;
        -k)
            COMPREPLY=( $(compgen -W "$(seq 1 65535)" -- ${cur}) )
            return 0
            ;;
        -p)
            COMPREPLY=( $(compgen -W "$(seq 1 65535)" -- ${cur}) )
            return 0
            ;;
    esac
}

complete -F _miniServer miniServer
