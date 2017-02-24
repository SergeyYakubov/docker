#!/bin/bash

mkdir $HOME

if [[ -z $1 ]]; then
	bash
else
	$*
fi

