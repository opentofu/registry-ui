import { useState } from "react";
import { Header } from "../../components/Header";
import { Paragraph } from "../../components/Paragraph";
import PatternBg from "../../components/PatternBg";
import { Footer } from "@/components/Footer";
import { MetaTags } from "@/components/MetaTags";
import { SearchExperience } from "./SearchExperience";
import clsx from "clsx";

export function Home() {
  return (
    <>
      <MetaTags />
      <SearchExperience />
    </>
  );
}